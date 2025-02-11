package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	registryMapping map[string]string
	mappingMutex    sync.RWMutex
	configPath      = "/etc/webhook/config/registry-mappings"
)

func loadConfigFromFile() error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Split the file content into lines
	lines := strings.Split(string(data), "\n")

	mappingMutex.Lock()
	defer mappingMutex.Unlock()

	// Clear existing mappings
	registryMapping = make(map[string]string)

	// Parse each line as key=value
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			registryMapping[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return nil
}

func watchConfig() {
	// Get watch interval from environment variable, default to 600 seconds (10 minutes)
	watchInterval := 3600
	if intervalStr := os.Getenv("CONFIG_WATCH_INTERVAL"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil {
			watchInterval = interval
		} else {
			log.Printf("Invalid CONFIG_WATCH_INTERVAL value: %v, using default", err)
		}
	}

	for {
		if err := loadConfigFromFile(); err != nil {
			log.Printf("Error reloading config: %v", err)
		} else {
			log.Printf("Successfully reloaded registry mappings")
		}
		time.Sleep(time.Duration(watchInterval) * time.Second)
	}
}

func mutateImage(image string) (string, bool) {
	mappingMutex.RLock()
	defer mappingMutex.RUnlock()

	for src, dst := range registryMapping {
		// Check for explicit prefix
		if strings.HasPrefix(image, src+"/") {
			newImage := strings.Replace(image, src+"/", dst+"/", 1)
			return newImage, true
		}

		// Handle implicit docker.io references
		if src == "docker.io" && !strings.Contains(image, "/") {
			return dst + "/" + image, true
		}
	}
	return image, false
}

func generatePatches(pod *corev1.Pod) ([]map[string]interface{}, error) {
	var patches []map[string]interface{}

	// Handle init containers
	for i, container := range pod.Spec.InitContainers {
		if newImage, changed := mutateImage(container.Image); changed {
			patch := map[string]interface{}{
				"op":    "replace",
				"path":  fmt.Sprintf("/spec/initContainers/%d/image", i),
				"value": newImage,
			}
			patches = append(patches, patch)
		}
	}

	// Handle regular containers
	for i, container := range pod.Spec.Containers {
		log.Printf("Generate Patches container image: %v ", container.Image)
		if newImage, changed := mutateImage(container.Image); changed {
			patch := map[string]interface{}{
				"op":    "replace",
				"path":  fmt.Sprintf("/spec/containers/%d/image", i),
				"value": newImage,
			}
			patches = append(patches, patch)
		}
	}

	return patches, nil
}

func handleMutation(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	var admissionReviewReq admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal request: %v", err), http.StatusBadRequest)
		return
	}

	// Parse Pod object
	pod := corev1.Pod{}
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod); err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal pod: %v", err), http.StatusBadRequest)
		return
	}

	// Generate patches for container images
	patches, err := generatePatches(&pod)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate patches: %v", err), http.StatusInternalServerError)
		return
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal patches: %v", err), http.StatusInternalServerError)
		return
	}

	admissionReviewResp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     admissionReviewReq.Request.UID,
			Allowed: true,
			Patch:   patchBytes,
			PatchType: func() *admissionv1.PatchType {
				pt := admissionv1.PatchTypeJSONPatch
				return &pt
			}(),
		},
	}

	respBytes, err := json.Marshal(admissionReviewResp)
	if err != nil {
		log.Printf("failed to marshal response for logging: %v", err)
	} else {
		log.Printf("Response body: %s", string(respBytes))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(admissionReviewResp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	// Initial config load
	if err := loadConfigFromFile(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Start config watcher in background
	go watchConfig()

	http.HandleFunc("/mutate", handleMutation)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Starting webhook server on :8443")
	if err := http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
