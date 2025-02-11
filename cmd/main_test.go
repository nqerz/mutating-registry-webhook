package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func TestMutateImage(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name     string
		image    string
		mappings map[string]string
		want     string
		changed  bool
	}{
		{
			name:  "should replace registry",
			image: "docker.io/nginx:latest",
			mappings: map[string]string{
				"docker.io": "private.registry.com",
			},
			want:    "private.registry.com/nginx:latest",
			changed: true,
		},
		{
			name:  "should replace implicit docker.io registry",
			image: "nginx:latest",
			mappings: map[string]string{
				"docker.io": "private.registry.com",
			},
			want:    "private.registry.com/nginx:latest",
			changed: true,
		},
		{
			name:  "should not replace registry",
			image: "quay.io/nginx:latest",
			mappings: map[string]string{
				"docker.io": "private.registry.com",
			},
			want:    "quay.io/nginx:latest",
			changed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the global registry mapping
			registryMapping = tt.mappings

			// Test the mutation
			got, changed := mutateImage(tt.image)
			if got != tt.want {
				t.Errorf("mutateImage() got = %v, want %v", got, tt.want)
			}
			if changed != tt.changed {
				t.Errorf("mutateImage() changed = %v, want %v", changed, tt.changed)
			}
		})
	}
}

func TestGeneratePatches(t *testing.T) {
	// Setup test pod
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{Name: "init", Image: "docker.io/busybox:latest"},
			},
			Containers: []corev1.Container{
				{Name: "app", Image: "docker.io/nginx:latest"},
			},
		},
	}

	// Setup registry mapping
	registryMapping = map[string]string{
		"docker.io": "private.registry.com",
	}

	// Generate patches
	patches, err := generatePatches(pod)
	if err != nil {
		t.Fatalf("generatePatches() error = %v", err)
	}

	// Verify patches
	expectedPatches := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/spec/initContainers/0/image",
			"value": "private.registry.com/busybox:latest",
		},
		{
			"op":    "replace",
			"path":  "/spec/containers/0/image",
			"value": "private.registry.com/nginx:latest",
		},
	}

	if len(patches) != len(expectedPatches) {
		t.Fatalf("generatePatches() got %d patches, want %d", len(patches), len(expectedPatches))
	}

	// Compare patches
	for i, patch := range patches {
		expected := expectedPatches[i]
		if patch["op"] != expected["op"] || patch["path"] != expected["path"] || patch["value"] != expected["value"] {
			t.Errorf("patch[%d] = %v, want %v", i, patch, expected)
		}
	}
}

func TestHandleMutation(t *testing.T) {
	// Setup registry mapping
	registryMapping = map[string]string{
		"docker.io": "private.registry.com",
	}

	// Create a test pod
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app", Image: "docker.io/nginx:latest"},
			},
		},
	}

	// Create admission review request
	podBytes, err := json.Marshal(pod)
	if err != nil {
		t.Fatalf("failed to marshal pod: %v", err)
	}

	review := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "test-uid",
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}

	// Marshal the admission review
	reviewBytes, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("failed to marshal admission review: %v", err)
	}

	// Create request
	req := httptest.NewRequest("POST", "/mutate", bytes.NewReader(reviewBytes))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handleMutation(rr, req)

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Parse response
	var responseReview admissionv1.AdmissionReview
	if err := json.NewDecoder(rr.Body).Decode(&responseReview); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify response
	if responseReview.Response.UID != review.Request.UID {
		t.Errorf("handler returned wrong UID: got %v want %v", responseReview.Response.UID, review.Request.UID)
	}

	if !responseReview.Response.Allowed {
		t.Error("handler returned not allowed")
	}

	if responseReview.Response.PatchType == nil || *responseReview.Response.PatchType != admissionv1.PatchTypeJSONPatch {
		t.Error("handler returned wrong patch type")
	}

	// Verify patches
	var patches []map[string]interface{}
	if err := json.Unmarshal(responseReview.Response.Patch, &patches); err != nil {
		t.Fatalf("failed to unmarshal patches: %v", err)
	}

	expectedPatch := map[string]interface{}{
		"op":    "replace",
		"path":  "/spec/containers/0/image",
		"value": "private.registry.com/nginx:latest",
	}

	if len(patches) != 1 {
		t.Fatalf("wrong number of patches: got %d want 1", len(patches))
	}

	if patches[0]["op"] != expectedPatch["op"] ||
		patches[0]["path"] != expectedPatch["path"] ||
		patches[0]["value"] != expectedPatch["value"] {
		t.Errorf("wrong patch: got %v want %v", patches[0], expectedPatch)
	}
}

func TestHealthHandler(t *testing.T) {
	// Create request
	req := httptest.NewRequest("GET", "/health", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	healthHandler(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Check response body
	if rr.Body.String() != "ok" {
		t.Errorf("handler returned wrong body: got %v want %v", rr.Body.String(), "ok")
	}
}
