#!/bin/bash

# Default namespace if not provided
NAMESPACE=${1:-"default"}

set -e

# Create temporary and target directories
TMP_DIR=".tmp"
mkdir -p "${TMP_DIR}"

# Generate CA key and certificate
openssl genrsa -out "${TMP_DIR}/ca.key" 2048
openssl req -x509 -new -nodes -key "${TMP_DIR}/ca.key" -days 365 -out "${TMP_DIR}/ca.crt" -subj "/CN=Webhook CA"

# Generate server key and certificate signing request
openssl genrsa -out "${TMP_DIR}/tls.key" 2048
openssl req -new -key "${TMP_DIR}/tls.key" -out "${TMP_DIR}/tls.csr" -subj "/CN=mutating-registry-webhook.${NAMESPACE}.svc" -config <(
cat <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = mutating-registry-webhook.${NAMESPACE}.svc
DNS.2 = mutating-registry-webhook.${NAMESPACE}.svc.cluster.local
EOF
)

# Sign the certificate
openssl x509 -req -in "${TMP_DIR}/tls.csr" -CA "${TMP_DIR}/ca.crt" -CAkey "${TMP_DIR}/ca.key" -CAcreateserial -out "${TMP_DIR}/tls.crt" -days 365 -extensions v3_req -extfile <(
cat <<EOF
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = mutating-registry-webhook.${NAMESPACE}.svc
DNS.2 = mutating-registry-webhook.${NAMESPACE}.svc.cluster.local
EOF
)

# Create Kubernetes secret with both TLS and CA certificates
cat > secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: mutating-registry-webhook-tls
  namespace: ${NAMESPACE}
type: kubernetes.io/tls
data:
  tls.crt: $(base64 -w0 < "${TMP_DIR}/tls.crt")
  tls.key: $(base64 -w0 < "${TMP_DIR}/tls.key")
  ca.crt: $(base64 -w0 < "${TMP_DIR}/ca.crt")
EOF

# Cleanup
rm -rf "${TMP_DIR}"