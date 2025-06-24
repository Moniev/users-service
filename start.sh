#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

LOG() { printf "${BLUE}[LOG]${NC} %s\n" "$1"; }
OK() { printf "${GREEN}[OK]${NC} %s\n" "$1"; }
WRN() { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
ERR() { printf "${RED}[ERROR]${NC} %s\n" "$1"; exit 1; }

check_command() {
    if ! command -v "$1" &> /dev/null; then
        ERR "'$1' command is not installed. Please install it to proceed."
    fi
}

check_openssl_version() {
    LOG "Checking OpenSSL version..."
    OPENSSL_VERSION=$(openssl version | awk '{print $2}')
    OPENSSL_MAJOR=$(echo "$OPENSSL_VERSION" | cut -d. -f1)
    OPENSSL_MINOR=$(echo "$OPENSSL_VERSION" | cut -d. -f2)
    if [ "$OPENSSL_MAJOR" -lt 1 ] || { [ "$OPENSSL_MAJOR" -eq 1 ] && [ "$OPENSSL_MINOR" -lt 1 ]; }; then
        ERR "OpenSSL version $OPENSSL_VERSION is too old. Version 1.1.0 or higher is required."
    fi
    OK "OpenSSL version $OPENSSL_VERSION is compatible"
}

check_environment() {
    LOG "Checking for required tools and environment..."
    check_command "docker"
    check_command "kubectl"
    check_command "kubeseal"
    check_command "jq"

    if ! docker info >/dev/null 2>&1; then ERR "Docker is not running. Please start it."; fi
    if ! kubectl get nodes >/dev/null 2>&1; then ERR "Kubernetes is not responding. Please check your cluster and kubeconfig."; fi
    OK "Docker, Kubernetes, and required tools are ready."
}

install_sealed_secrets() {
    LOG "Installing or updating Sealed Secrets controller..."
    kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.26.1/controller.yaml
    WRN "Waiting for Sealed Secrets controller to become ready..."
    kubectl wait --for=condition=Available --timeout=300s deployment/sealed-secrets-controller -n kube-system || ERR "Sealed Secrets controller did not become ready in time."
    OK "Sealed Secrets controller is running."
}

install_cert_manager() {
    LOG "Installing or updating cert-manager..."
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.14.5/cert-manager.yaml
    WRN "Waiting for cert-manager webhook to become ready. This might take a minute..."
    kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager-webhook -n cert-manager || ERR "cert-manager webhook did not become ready in time."
    OK "cert-manager is running."
}

create_sealed_secret() {
    local secret_name="$1"
    local namespace="$2"
    shift 2

    LOG "Creating and sealing secret '$secret_name' in namespace '$namespace'..."
    kubectl delete secret "$secret_name" -n "$namespace" --ignore-not-found=true

    if ! kubectl create secret generic "$secret_name" --namespace "$namespace" "${@}" --dry-run=client -o yaml | \
        kubeseal --controller-name sealed-secrets-controller --controller-namespace kube-system --format yaml --allow-empty-data | \
        kubectl apply -f -; then
        ERR "Failed to create or apply sealed secret '$secret_name'."
    fi
    OK "SealedSecret for '$secret_name' applied."
}

wait_for_all_secrets() {
    LOG "Waiting for all secrets to be unsealed by the controller..."
    local secrets_to_check=("$@")
    local all_ready=false
    local retries=30
    local interval=5

    while [ $retries -gt 0 ] && [ "$all_ready" = false ]; do
        all_ready=true
        local unready_secrets=""

        for secret_ref in "${secrets_to_check[@]}"; do
            local namespace=$(echo "$secret_ref" | cut -d'/' -f1)
            local secret_name=$(echo "$secret_ref" | cut -d'/' -f2)
            
            if ! kubectl get secret "$secret_name" -n "$namespace" &> /dev/null; then
                all_ready=false
                unready_secrets+="$secret_ref "
            fi
        done

        if [ "$all_ready" = false ]; then
            WRN "Still waiting for secrets: $unready_secrets ($retries retries left)"
            sleep $interval
            retries=$((retries - 1))
        fi
    done

    if [ "$all_ready" = false ]; then
        ERR "Timed out waiting for secrets to be created. Check the sealed-secrets-controller logs."
    fi

    OK "All secrets successfully unsealed and are available."
}


generate_and_setup_ca_issuer() {
    LOG "Setting up the main Certificate Authority and ClusterIssuer..."
    local ca_dir="k8s/ca"
    mkdir -p "$ca_dir"

    local ca_key_path="$ca_dir/ca.key"
    local ca_cert_path="$ca_dir/ca.crt"

    if [ ! -f "$ca_key_path" ]; then
        LOG "Generating new root CA key and certificate..."
        openssl genrsa -out "$ca_key_path" 4096 >/dev/null 2>&1 || ERR "Failed to generate CA key"
        openssl req -x509 -new -nodes -key "$ca_key_path" -sha256 -days 3650 \
                -out "$ca_cert_path" -subj "/CN=factory-chainline-cluster-ca" -batch >/dev/null 2>&1 || ERR "Failed to generate CA certificate"
        OK "Root CA generated."
    else
        OK "Root CA already exists. Skipping generation."
    fi

    LOG "Sealing the root CA key pair into a secret for cert-manager..."
    create_sealed_secret "factory-chainline-ca-secret" "cert-manager" \
        "--from-file=tls.crt=$ca_cert_path" \
        "--from-file=tls.key=$ca_key_path"

    LOG "Creating ClusterIssuer 'factory-chainline-ca-issuer'..."
    cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: factory-chainline-ca-issuer
spec:
  ca:
    secretName: factory-chainline-ca-secret
EOF
    OK "ClusterIssuer 'factory-chainline-ca-issuer' created successfully."
}

create_static_secrets() {
    LOG "Creating static application secrets..."
        
    create_sealed_secret "users-service-secrets" "users-service" \
        --from-literal=SMTP_PASSWORD="eocq irvk nyke sxdy" \
        --from-literal=ABUSEIPDB_KEY="b2b7d0295ce08f57f84e940fee9ce1b63c2d3b1755fe5fb04e67e75e486e6ff62b54749c2f8d7e32" \
        --from-literal=HASHING_PEPPER="PxHxLf3x2ziqlRNhjY9jgQJl6pA69wtJf6NtZAfVNE8HvkeDc3LoxlgM9xuklWPHarCc2EYS1c6gR1oPVQ0JD4CursyAEC14EDBBS8r0BE0aFmvowfOdmykwJPPhIaab" \
        --from-literal=TWILIO_ID="AC5652227976e14c03c76c6bc7bbf6ab9e" \
        --from-literal=TWILIO_PASSWORD="AC5652227976e14c03c76c6bc7bbf6ab9e" \
        --from-literal=DB_USER="2Atml0xquH6hYOGz0rjzPYL00Td86Kl7fNx9mWmy4Kg2ulT7WP" \
        --from-literal=DB_PASSWORD="4j2Q0eZk22tZDDJD2Pw2YCZKYKfmn97YJ1nU2h9shsg45B0dL0" \
        --from-literal=DB_NAME="MAIN_DB" \
        --from-literal=DB_PORT=5432 \
        --from-literal=DB_HOST="postgres.postgres.svc.cluster.local" \
        --from-literal=REDIS_PASSWORD="5IDLQZaWef67kC0vT83KAeWg2NbnBWOapp4pp5mFJYUQ70Sn8a" \
        --from-literal=REDIS_SECRET_KEY="CJl5oPQ3Up4XOY0qxB9R12fDH2SH2d0u" \

    if [ ! -f "k8s/certificates/private_key.pem" ] || [ ! -f "k8s/certificates/public_key.pem" ]; then
        ERR "private_key.pem or public_key.pem not found. Please generate them first."
    fi
    create_sealed_secret "ed25519-keys-secret" "users-service" \
        "--from-file=private_key.pem=./k8s/certificates/private_key.pem" \
        "--from-file=public_key.pem=./k8s/certificates/public_key.pem"


    OK "All static secrets have been sealed and applied."
}

check_openssl_version
check_environment
install_sealed_secrets
install_cert_manager

LOG "Creating application namespaces..."
NAMESPACES=(cert-manager argocd)
for ns in "${NAMESPACES[@]}"; do
    kubectl create namespace "$ns" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
done
OK "All namespaces created or already exist."

generate_and_setup_ca_issuer
create_static_secrets

ALL_SECRETS=(
    "cert-manager/factory-chainline-ca-secret"
    "grafana/grafana-admin-credentials"
    "users-service/users-service-secrets"
    "users-service/ed25519-keys-secret" 
)
wait_for_all_secrets "${ALL_SECRETS[@]}"

OK "Deployment tasks completed!"
LOG "It may take a few minutes for all pods to become ready and for cert-manager to issue all certificates."
LOG "Check status with: ${CYAN}kubectl get pods,certificates,clusterissuers --all-namespaces${NC}"

if [ "$1" = "cleanup" ]; then
    LOG "Cleanup logic to be implemented."
else
    LOG "Skipping cleanup of certificate directories. Use '$0 cleanup' to remove them after successful deployment and verification."
fi

LOG "Check pod status with: ${CYAN}kubectl get pods --all-namespaces${NC}"
LOG "Check pod logs using 'kubectl logs <pod-name> -n <namespace>'"