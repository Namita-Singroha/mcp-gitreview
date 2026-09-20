#!/bin/bash
# install-runtime.sh - Deploy MaaS Runtime Platform
# Usage: ./install-runtime.sh <values-file> [cluster-override-file]
#
# Arguments:
#   values-file           Base values file (required)
#   cluster-override-file Optional second values file — merged on top of base.
#                         Use this for cluster-specific overrides such as
#                         disabling cert-manager/Keycloak that are pre-installed.
# Example:
#   ./install-runtime.sh \
#     examples/Fusion-Agentic-Assistance-Platform/values.yaml \
#     examples/my-cluster-values.yaml

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHARTS_DIR="$PROJECT_ROOT/deploy/helm"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Positional arguments
VALUES_FILE="${1:-$PROJECT_ROOT/examples/Fusion-Agentic-Assistance-Platform/values.yaml}"
OVERRIDE_FILE="${2:-}"   # optional cluster-specific override (second argument)

OPERATORS_VALUES_FILE="${OPERATORS_VALUES_FILE:-$VALUES_FILE}"
PLATFORM_VALUES_FILE="${PLATFORM_VALUES_FILE:-$VALUES_FILE}"
RUNTIME_VALUES_FILE="${RUNTIME_VALUES_FILE:-$VALUES_FILE}"

# Build the -f flag string used in every helm call.
# If an override file was supplied it is appended as a second -f so its values
# win over the base file (standard Helm merge order).
HELM_VALUES_ARGS="-f $VALUES_FILE"
if [ -n "$OVERRIDE_FILE" ]; then
    HELM_VALUES_ARGS="$HELM_VALUES_ARGS -f $OVERRIDE_FILE"
fi

echo -e "${GREEN}=== MaaS Runtime Installation ===${NC}"
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check if oc is installed
if ! command -v oc &> /dev/null; then
    echo -e "${RED}Error: OpenShift CLI (oc) is not installed${NC}"
    exit 1
fi

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo -e "${RED}Error: Helm is not installed${NC}"
    exit 1
fi

# Check if logged into cluster
if ! oc whoami &> /dev/null; then
    echo -e "${RED}Error: Not logged into OpenShift cluster${NC}"
    echo "Please run: oc login <cluster-url>"
    exit 1
fi

# Check if user is cluster admin
if ! oc auth can-i '*' '*' --all-namespaces &> /dev/null; then
    echo -e "${YELLOW}Warning: You may not have cluster-admin privileges${NC}"
    echo "Some operations may fail. Continue? (y/n)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# ── GPU Readiness Check ───────────────────────────────────────────────────────
# Queries node.status.allocatable — NOT node labels.
# amd.com/gpu is simultaneously a node label AND a resource name; only
# .status.allocatable confirms a GPU is actually schedulable.
echo "Checking GPU resources on cluster nodes..."
NVIDIA_GPU_NODES=$(oc get nodes \
    -o jsonpath='{range .items[*]}{.status.allocatable.nvidia\.com/gpu}{"\n"}{end}' \
    2>/dev/null | grep -v '^$' | grep -v '^0$' | wc -l | tr -d ' ')
AMD_GPU_NODES=$(oc get nodes \
    -o jsonpath='{range .items[*]}{.status.allocatable.amd\.com/gpu}{"\n"}{end}' \
    2>/dev/null | grep -v '^$' | grep -v '^0$' | wc -l | tr -d ' ')

if [[ "$NVIDIA_GPU_NODES" -gt 0 && "$AMD_GPU_NODES" -gt 0 ]]; then
    echo -e "${GREEN}✓ Mixed GPU cluster detected${NC}"
    echo "  NVIDIA: $NVIDIA_GPU_NODES node(s) with nvidia.com/gpu allocatable"
    echo "  AMD:    $AMD_GPU_NODES node(s) with amd.com/gpu allocatable"
    echo "  Use vendor-specific overlays when deploying models (accelerator.vendor: nvidia|amd)"
elif [[ "$NVIDIA_GPU_NODES" -gt 0 ]]; then
    echo -e "${GREEN}✓ NVIDIA GPU cluster: $NVIDIA_GPU_NODES node(s) with nvidia.com/gpu allocatable${NC}"
elif [[ "$AMD_GPU_NODES" -gt 0 ]]; then
    echo -e "${GREEN}✓ AMD GPU cluster: $AMD_GPU_NODES node(s) with amd.com/gpu allocatable${NC}"
else
    echo -e "${YELLOW}⚠ No GPU resources (nvidia.com/gpu or amd.com/gpu) allocatable on any node.${NC}"
    echo "  MaaS operators will install. GPU-backed model serving requires GPU operators from isf-compute-operator."
fi
echo ""

# Check if values files exist
for file in "$VALUES_FILE" "$OPERATORS_VALUES_FILE" "$PLATFORM_VALUES_FILE" "$RUNTIME_VALUES_FILE"; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}Error: Values file not found: $file${NC}"
        exit 1
    fi
done
if [ -n "$OVERRIDE_FILE" ] && [ ! -f "$OVERRIDE_FILE" ]; then
    echo -e "${RED}Error: Override file not found: $OVERRIDE_FILE${NC}"
    exit 1
fi
if [ -n "$OVERRIDE_FILE" ]; then
    echo "  Override file: $OVERRIDE_FILE"
fi

echo -e "${GREEN}✓ Prerequisites check passed${NC}"
echo ""

# Check for default StorageClass
echo "Checking for default StorageClass..."
if ! oc get storageclass -o jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}' | grep -q .; then
    echo -e "${RED}Error: No default StorageClass found${NC}"
    echo "Please configure a default StorageClass before proceeding"
    exit 1
fi
echo -e "${GREEN}✓ Default StorageClass found${NC}"
echo ""

# Prompt for passwords if using Keycloak.
# Check base file first, then let override file win (last file sets final value).
_keycloak_enabled() {
    local enabled="false"
    # Walk each values file in order; last explicit setting wins.
    for f in "$VALUES_FILE" ${OVERRIDE_FILE:+"$OVERRIDE_FILE"}; do
        if grep -A 3 "keycloak:" "$f" 2>/dev/null | grep -q "enabled: true"; then
            enabled="true"
        elif grep -A 3 "keycloak:" "$f" 2>/dev/null | grep -q "enabled: false"; then
            enabled="false"
        fi
    done
    echo "$enabled"
}

if [ "$(_keycloak_enabled)" = "true" ]; then
    echo "Keycloak is enabled. Please provide passwords:"
    echo ""
    
    if [ -z "$ADMIN_PASSWORD" ]; then
        read -rsp "Enter admin password: " ADMIN_PASSWORD
        echo ""
    fi
    
    if [ -z "$USER_PASSWORD" ]; then
        read -rsp "Enter user password: " USER_PASSWORD
        echo ""
    fi
    
    KEYCLOAK_ARGS="--set authentication.keycloak.realm.admin.password=$ADMIN_PASSWORD --set authentication.keycloak.realm.user.password=$USER_PASSWORD"
else
    KEYCLOAK_ARGS=""
fi

echo ""
echo -e "${GREEN}=== Phase 1: Installing Dependency Operator Subscriptions ===${NC}"
echo "Values file: $OPERATORS_VALUES_FILE"
echo "Chart: $CHARTS_DIR/maas-operators"
echo ""

helm upgrade --install maas-operators "$CHARTS_DIR/maas-operators" \
    $HELM_VALUES_ARGS \
    --timeout 20m \
    --wait

echo ""
echo -e "${GREEN}✓ Dependency operator subscriptions installed${NC}"
echo ""

# Wait for OpenShift AI operator to be ready
echo "Waiting for OpenShift AI operator to be ready..."
for i in {1..60}; do
    if oc get csv -n redhat-ods-operator 2>/dev/null | grep -q "Succeeded" || oc get deployment -n redhat-ods-operator 2>/dev/null | grep -q "rhods\|opendatahub"; then
        echo -e "${GREEN}✓ OpenShift AI operator ready${NC}"
        break
    fi
    sleep 5
done

# Wait for DataScienceCluster CRD to be available
echo "Waiting for DataScienceCluster CRD to be available..."
for i in {1..60}; do
    if oc get crd datascienceclusters.datasciencecluster.opendatahub.io &>/dev/null; then
        echo -e "${GREEN}✓ DataScienceCluster CRD available${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo -e "${RED}✗ DataScienceCluster CRD not available after 5 minutes${NC}"
        exit 1
    fi
    sleep 5
done

# Check Kuadrant CRD (skip if disabled or missing)
if oc get crd kuadrants.kuadrant.io &>/dev/null; then
    echo -e "${GREEN}✓ Kuadrant CRD available${NC}"
else
    echo -e "${YELLOW}ℹ Kuadrant CRD not found (optional/disabled — skipping)${NC}"
fi

# Check LeaderWorkerSet CRD (skip if disabled or missing)
if oc get crd leaderworkersetoperators.operator.openshift.io &>/dev/null || oc get crd leaderworkersets.leaderworkerset.x-k8s.io &>/dev/null; then
    echo -e "${GREEN}✓ LeaderWorkerSet CRD available${NC}"
else
    echo -e "${YELLOW}ℹ LeaderWorkerSet CRD not found (optional/disabled — skipping)${NC}"
fi

echo ""
echo -e "${GREEN}=== Phase 2: Creating DataScienceCluster and Operator Instances ===${NC}"
echo "Values file: $PLATFORM_VALUES_FILE"
echo "Chart: $CHARTS_DIR/maas-platform"
echo ""

helm upgrade --install maas-platform "$CHARTS_DIR/maas-platform" \
    $HELM_VALUES_ARGS \
    --timeout 20m \
    --wait

echo ""
echo "Waiting for DataScienceCluster to be ready..."
if oc wait --for=condition=Ready datasciencecluster default-dsc --timeout=15m 2>/dev/null; then
    echo -e "${GREEN}✓ DataScienceCluster ready${NC}"
else
    echo -e "${RED}✗ DataScienceCluster not ready${NC}"
    echo "Please check the DataScienceCluster status and try again"
    exit 1
fi

echo ""
echo -e "${GREEN}=== Phase 3: Installing MaaS Runtime Resources ===${NC}"
echo "Values file: $RUNTIME_VALUES_FILE"
echo "Chart: $CHARTS_DIR/maas-runtime"
echo ""

# Install or upgrade the main runtime resources
echo "Installing MaaS runtime resources (gateway, model registry, workbench storage, etc.)..."
helm upgrade --install maas-runtime "$CHARTS_DIR/maas-runtime" \
    $HELM_VALUES_ARGS \
    $KEYCLOAK_ARGS \
    --timeout 10m \
    --force

echo ""
echo -e "${GREEN}✓ MaaS Runtime resources installation complete${NC}"
echo ""

# Wait for additional components
echo "Waiting for additional components to be ready..."
echo ""

# Wait for Kuadrant
echo "Waiting for Kuadrant..."
if oc wait --for=condition=Ready kuadrant kuadrant -n kuadrant-system --timeout=5m 2>/dev/null; then
    echo -e "${GREEN}✓ Kuadrant ready${NC}"
else
    echo -e "${YELLOW}⚠ Kuadrant not ready yet${NC}"
fi

# Wait for Keycloak if enabled
if [ -n "$KEYCLOAK_ARGS" ]; then
    echo "Waiting for Keycloak..."
    if oc wait --for=condition=Ready keycloak keycloak -n keycloak --timeout=10m 2>/dev/null; then
        echo -e "${GREEN}✓ Keycloak ready${NC}"
    else
        echo -e "${YELLOW}⚠ Keycloak not ready yet${NC}"
    fi
fi

echo ""
echo -e "${GREEN}=== Installation Summary ===${NC}"
echo ""
echo "MaaS Runtime has been deployed!"
echo ""
echo "Next steps:"
echo "1. Deploy models using: ./deploy-model.sh <model-values-file>"
echo "2. Check status: oc get all -n maas-models"
echo "3. View logs: oc logs -n maas-models -l app.kubernetes.io/component=model-service"
echo ""

# Display useful URLs
echo "Useful URLs:"
CONSOLE_URL=$(oc whoami --show-console 2>/dev/null || echo "N/A")
echo "  OpenShift Console: $CONSOLE_URL"

if [ -n "$KEYCLOAK_ARGS" ]; then
    KEYCLOAK_URL=$(oc get route -n keycloak keycloak -o jsonpath='{.spec.host}' 2>/dev/null || echo "N/A")
    echo "  Keycloak: https://$KEYCLOAK_URL"
fi

GRAFANA_URL=$(oc get route -n grafana grafana-route -o jsonpath='{.spec.host}' 2>/dev/null || echo "N/A")
if [ "$GRAFANA_URL" != "N/A" ]; then
    echo "  Grafana: https://$GRAFANA_URL"
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"

