# Quickstart: CPU-Based LLM Inference on IBM Fusion (GitOps Deployment)

Organizations evaluating generative AI need a fast, cost-effective path to running open-source language models on existing x86 infrastructure, without GPU procurement cycles. CPU inference via vLLM lets you stand up real LLM endpoints on standard compute nodes in hours, using the same GitOps tooling used across IBM Fusion AI workloads.

This quickstart demonstrates how to deploy multiple CPU-based language models on IBM Fusion using Red Hat OpenShift GitOps (ArgoCD). Models run on x86 CPU using the vLLM CPU `ServingRuntime` for KServe, with model artifacts stored in IBM Fusion's OpenShift Data Foundation (ODF) object storage. All resources (namespaces, runtimes, inference services, and secrets) are declared in Git and continuously reconciled by ArgoCD.

This guide is intended for platform engineers, AI infrastructure teams, and developers who want to learn the IBM Fusion AI deployment model, evaluate open-source models, and gain hands-on experience with OpenShift AI, KServe, vLLM, and GitOps-based model deployment. CPU-based inference provides a low-cost environment for development, validation, and functional testing before adopting GPU-based MaaS deployments for production-scale workloads.

> **CPU vs GPU: scope of this guide.** CPU inference uses the standard KServe `ServingRuntime + InferenceService` path. It does **not** integrate with the MaaS gateway, subscriptions, or API-key enforcement. Those capabilities require a GPU (`LLMInferenceService`) runtime. For the GPU/MaaS GitOps deployment, see [Quickstart: IBM Fusion Model-as-a-Service Platform (GitOps Deployment and Customization)](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/infoDocs/gitops-deployment-guide_updated.md).

By the end of this guide, you will have GitOps-managed CPU LLM inference services running on IBM Fusion, accessible through secure OpenShift Routes with bearer-token authentication.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [What You'll Build](#what-youll-build)
- [Prerequisites](#prerequisites)
- [Deployment Steps](#deployment-steps)
  - [Step 1: Fork and Clone the Repository](#step-1-fork-and-clone-the-repository)
  - [Step 2: Verify RHOAI and Model Registration](#step-2-verify-rhoai-and-model-registration)
  - [Step 3: Update repoURL and targetRevision](#step-3-update-repourl-and-targetrevision)
  - [Step 4: Deploy Model AppProjects and Applications](#step-4-deploy-model-appprojects-and-applications)
  - [Step 5: Sync the Model Applications](#step-5-sync-the-model-applications)
  - [Step 6: Monitor Model Pods](#step-6-monitor-model-pods)
  - [Step 7: Test Inference Endpoints](#step-7-test-inference-endpoints)
- [What's Deployed](#whats-deployed)
- [Add a New CPU Model](#add-a-new-cpu-model)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)
- [Summary](#summary)

---

## Architecture Overview

> **The primary goal of this quickstart is to demonstrate OpenShift AI model serving workflows on IBM Fusion using a low-cost CPU footprint.** This guide relies on OpenShift AI (RHOAI) for model lifecycle management, KServe model serving, Model Registry integration, and `ServingRuntime`/`InferenceService` resources. ArgoCD applies the manifests, but OpenShift AI's admission webhooks, model serving controllers, and KServe integration are what schedule, monitor, and serve the models. The diagram below labels the open-source components (KServe, vLLM); these are managed by OpenShift AI.

```
┌──────────────────────────────────────────────────────────────────────┐
│       IBM Fusion: CPU LLM Inference (GitOps)                         │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   Developer / Platform Engineer                                      │
│          │                                                           │
│          │  git commit + push (model values / ArgoCD manifests)      │
│          ▼                                                           │
│   ┌──────────────────────────────────┐   ┌────────────────────────┐  │
│   │  Git Repository (Fusion-AI)      │   │  HashiCorp Vault       │  │
│   │  • ArgoCD Application manifests  │   │  • S3 credentials      │  │
│   │  • Helm chart + values files     │   │  (ESO syncs at deploy) │  │
│   │  • Per-model, per-env overrides  │   └────────────┬───────────┘  │
│   └──────────────┬───────────────────┘                │ ESO          │
│                  │  ArgoCD watches & reconciles       │ExternalSecret│
│                  ▼                                    ▼              │
│   ┌──────────────────────────────────────────────────────────┐       │
│   │       Red Hat OpenShift GitOps (ArgoCD)                  │       │
│   │       running on IBM Fusion cluster                      │       │
│   │  • One AppProject per environment                        │       │
│   │  • One Application per model                             │       │
│   │  • ignoreDifferences suppresses shared-resource drift    │       │
│   └───────────────┬──────────────────────────────────────────┘       │
│                   │  applies manifests via                           │
│                   ▼                                                  │
│   ┌──────────────────────────────────────────────────────────┐       │
│   │    Red Hat OpenShift AI (RHOAI) [platform operator]      │       │
│   │    • Model lifecycle management (Model Registry,         │       │
│   │      serving controller, admission webhook)              │       │
│   │    • Manages KServe model serving (open-source)          │       │
│   │    • Owns ServingRuntime / InferenceService CRDs         │       │
│   └───────────────┬──────────────────────────────────────────┘       │
│                   │  creates & reconciles resources in               │
│                   ▼                                                  │
│   ┌──────────────────────────────────────────────────────────┐       │
│   │         deploy-models-cpu  (Namespace)                   │       │
│   │                                                          │       │
│   │  ┌──────────────────┐  ┌─────────────────────────────┐   │       │
│   │  │  ServingRuntime  │  │  InferenceService           │   │       │
│   │  │  (vLLM CPU)      │◄─┤  (one per model)            │   │       │
│   │  │  per model       │  │  storage: ODF via KServe    │   │       │
│   │  └──────────────────┘  └──────────────┬──────────────┘   │       │
│   │                                        │                 │       │
│   │  Shared resources (resource-policy: keep):               │       │
│   │  • ServiceAccount  • ExternalSecret CRs                  │       │
│   │  • storage-config  • ClusterRole argocd-manager-vllm-cpu │       │
│   │                                        │ storage-        │       │
│   │                                        │ initializer     │       │
│   └────────────────────────────────────────┼─────────────────┘       │
│                                            │ downloads model         │
│                                            ▼ weights at pod start    │
│   ┌──────────────────────────────────────────────────────────┐       │
│   │      IBM Fusion Storage (OpenShift Data Foundation)      │       │
│   │         • ODF ObjectBucketClaim (auto-provisioned)       │       │
│   │         • model weights stored as S3 objects             │       │
│   └──────────────────────────────────────────────────────────┘       │
│                                                                      │
│   ┌──────────────────────────────────────────────────────────┐       │
│   │   OpenShift Route (TLS reencrypt, per model)             │       │
│   │   • Bearer-token auth (ServiceAccount token)             │       │
│   │   • Exposes /v1/chat/completions to external clients     │       │
│   └──────────────────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## What You'll Build

Three small CPU LLMs deployed as GitOps-managed KServe inference services on IBM Fusion:

| Model | Purpose |
|---|---|
| `Qwen2.5-1.5B-Instruct` | General-purpose instruction-following |
| `Qwen2.5-Coder-1.5B-Instruct` | Code generation and completion |
| `SmolLM2-1.7B-Instruct` | Compact general-purpose model |

All three share a namespace (`deploy-models-cpu`), S3 credentials via ESO/Vault, and a `ServiceAccount`. Each model gets its own `ServingRuntime`, `InferenceService`, and bearer-token-protected OpenShift Route.

> **CPU inference scope.** This guide uses the standard KServe `ServingRuntime + InferenceService` path with bearer-token (ServiceAccount) authentication. It does not integrate with the MaaS gateway, subscriptions, or API-key enforcement — those capabilities are available only for GPU-based deployments. See [What's Deployed](#whats-deployed) for the full resource inventory.

> **These models are examples, not requirements.** The same pattern works for any model whose weights can be served by vLLM on CPU.
---

## Prerequisites

> **OpenShift AI is required.** This guide relies on OpenShift AI (RHOAI) for model lifecycle management, KServe model serving, Model Registry integration, and `ServingRuntime`/`InferenceService` resources. OpenShift AI must be installed and in a `Ready` state before proceeding. See the [MaaS Quickstart README](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/README.md) for installation steps.

### Tested With

This guide has been tested and validated with the following software versions:

| Component | Version |
|---|---|
| Red Hat OpenShift | `4.19.9` |
| Red Hat OpenShift AI | `2.19.0` |
| vLLM CPU ServingRuntime | `0.6.x` |
| Red Hat OpenShift GitOps (ArgoCD) | `1.14+` |

### Required

- **Completed MaaS Platform Quickstart**: The following must already be in place before starting this guide — complete **Steps 1–4** of the [MaaS Quickstart README](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/README.md) if you have not done so:
  - ODF storage configured and an `ObjectBucketClaim` provisioned
  - Red Hat OpenShift AI (RHOAI) installed and in a `Ready` state
  - At least one model uploaded to ODF storage and registered in the Model Registry (its S3 artifact path becomes the `s3.modelPath` value in the Helm values file). Models can come from Hugging Face, a private registry, or any other source — the registry is source-agnostic. The three example models (`Qwen2.5-1.5B-Instruct`, `Qwen2.5-Coder-1.5B-Instruct`, `SmolLM2-1.7B-Instruct`) are pre-uploaded and registered; no action needed if you are using those.
  - For a full walkthrough of model registration and GitOps-based sync, see Steps 8–9 of the [MaaS GitOps Deployment Guide](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/infoDocs/gitops-deployment-guide_updated.md). To add a custom model, see [Add a New CPU Model](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/deploy/helm/model-deploy-vllm-cpu/README.md).
- **IBM Fusion cluster** with ODF and RHOAI operators healthy
- **Red Hat OpenShift 4.19.9 or later** with cluster-admin access
- **Red Hat OpenShift AI 3.4 or latest** installed and in a `Ready` state (see [MaaS Quickstart README](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/README.md))
- **Red Hat OpenShift GitOps operator** installed. This deploys ArgoCD into the `openshift-gitops` namespace. See [Quickstart: GitOps (ArgoCD) on IBM Fusion](https://community.ibm.com/community/user/blogs/christo-abraham/2026/05/27/fusion-gitops-quickstart) for installation steps
- **OpenShift CLI (`oc`)**: [Install oc](https://docs.openshift.com/container-platform/latest/cli_reference/openshift_cli/getting-started-cli.html)
- **ArgoCD CLI (`argocd`) v2.9+**: [Install argocd](https://argo-cd.readthedocs.io/en/stable/cli_installation/)
- **x86 (amd64) worker nodes**: vLLM CPU requires x86 architecture. No GPU required. Each model pod requires dedicated CPU and memory; ensure at least one worker node can satisfy the largest single-pod request. The three example models in this guide each require **4 vCPUs and 12 GiB memory** per pod:

  | Model | Parameters | Weights (bfloat16) | KV Cache (`VLLM_CPU_KVCACHE_SPACE`) | CPU request/limit | Memory request/limit |
  |---|---|---|---|---|---|
  | `Qwen2.5-1.5B-Instruct` | 1.5B | ~3 GiB | 6 GiB | 4 cores | 12 GiB |
  | `Qwen2.5-Coder-1.5B-Instruct` | 1.5B | ~3 GiB | 6 GiB | 4 cores | 12 GiB |
  | `SmolLM2-1.7B-Instruct` | 1.7B | ~3.4 GiB | 6 GiB | 4 cores | 12 GiB |

  All three models run in the same namespace (`deploy-models-cpu`). To run them concurrently, the cluster needs at least **12 vCPUs and 36 GiB of allocatable memory** across its worker nodes (3 × 4 vCPU + 3 × 12 GiB). A single large worker node or multiple smaller ones both satisfy this.

  > vLLM CPU performs an AOT (ahead-of-time) compilation warmup at startup. During this phase memory usage peaks above the steady-state serving footprint - **do not reduce memory limits below 12 GiB** for these models, as that will cause an OOMKill before the server is ready.

- **IBM Fusion with OpenShift Data Foundation (ODF)**: object storage is auto-provisioned via `ObjectBucketClaim`; no manual bucket or credential setup is required.
- **GitHub account** with write access to a fork of the [storage-fusion](https://github.com/IBM/storage-fusion) repository. The model deployment manifests live at `AI/quickstarts/model-as-a-service/` within the repository.

### Recommended

ESO with Vault is the default credential mode used in this guide. S3 credentials are stored in Vault and synced into the cluster by ESO at sync time, so no credentials are committed to Git or passed on the command line.

- **HashiCorp Vault**: stores S3 credentials outside Git. See [Deploying Vault Guide](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/fusion-gitops/docs/deploying-vault-guide.md)
- **External Secrets Operator (ESO)**: syncs secrets from Vault into the cluster at sync time. See [Deploying External Secrets Guide](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/fusion-gitops/docs/deploying-external-secrets-guide.md)

> **If Vault and ESO are not available**, you can disable ESO in the model values files and pass credentials manually at sync time instead. See the [Values Reference](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/deploy/helm/model-deploy-vllm-cpu/VALUES.md) for how to switch modes. Manual credentials are acceptable for development and evaluation environments.

### Verify Your Environment

```bash
# Check OpenShift version (should be 4.19.9+)
oc version
```

```bash
# Verify cluster-admin access
oc auth can-i '*' '*' --all-namespaces
```

```bash
# Check RHOAI is installed and Ready
oc get datasciencecluster default-dsc -o jsonpath='{.status.phase}'
```

**Expected output:**

```text
Ready
```

```bash
# Check OpenShift GitOps is running
oc get pods -n openshift-gitops | grep Running
```

```bash
# Verify ArgoCD CLI is configured
argocd version
```

```bash
# Confirm x86 nodes are available
oc get nodes -o custom-columns=NAME:.metadata.name,ARCH:.status.nodeInfo.architecture
```

---

## Deployment Steps

> **Environment note:** This guide uses the `prod` environment. For `dev` or `staging`, replace `prod` in all paths and namespace references. For example, use `environments/dev/` instead of `environments/prod/` and `deploy-models-cpu-dev` instead of `deploy-models-cpu`.
>
> **How ArgoCD renders resources:** Each model application renders the `model-deploy-vllm-cpu` Helm chart using a base `values.yaml` merged with a per-model override at `environments/prod/values-<model>-cpu.yaml`.

### Step 1: Fork and Clone the Repository

Fork the [IBM storage-fusion](https://github.com/IBM/storage-fusion) repository to your GitHub account, then clone the forked repository to your local environment.

```bash
git clone git@github.com:<your-username>/storage-fusion.git
cd storage-fusion/AI/quickstarts/model-as-a-service
```

The two directories you will work with:

| Directory | Purpose |
|---|---|
| `deploy/helm/model-deploy-vllm-cpu/` | Helm chart that templates all Kubernetes resources |
| `deploy/gitops/model-deploy-vllm-cpu/` | ArgoCD manifests including AppProject and Application per model |

---

### Step 2: Verify RHOAI and Model Registration

Confirm that the MaaS Platform Quickstart prerequisites are satisfied before proceeding. These checks correspond to the conditions listed under **Completed MaaS Platform Quickstart** in `## Prerequisites`.

**Check RHOAI is installed and Ready:**
```bash
oc get datasciencecluster default-dsc \
  -o jsonpath='{.status.phase}'
```
**Expected output:**
```text
Ready
```

**Check MaaS is enabled:**
```bash
oc get datasciencecluster default-dsc \
  -o jsonpath='{.spec.components.kserve.modelsAsService.managementState}'
```
**Expected output:**
```text
Managed
```

**Check models are registered in the Model Registry:**
```bash
oc get inferenceservice -n deploy-models-cpu
```
**Expected output:** one `InferenceService` row per model, all in `Ready` state.

**Check MaaSModelRef resources are Ready:**
```bash
oc get maasmodelref --all-namespaces
```
**Expected output:** one `MaaSModelRef` per model with `READY=True`.

> **Not ready?** Return to **Steps 1–4** of the [MaaS Quickstart README](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/README.md) to configure storage, install RHOAI, and register your models before continuing.

---

### Step 3: Update repoURL and targetRevision

Every ArgoCD Application manifest has a `source` block that tells ArgoCD where to pull the Helm chart from. Before applying anything, update these two fields across all Application manifests to point to your fork:

| Field | Purpose | Default |
|---|---|---|
| `repoURL` | Git repository containing the chart | `https://github.com/IBM/storage-fusion.git` |
| `targetRevision` | Branch, tag, or commit SHA to track | `master` (all environments) |

**Bulk update all environments at once** (run from the repo root):

```bash
# Update repoURL across all environments
find AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu \
  -name '*.yaml' | xargs sed -i '' \
  's|repoURL: https://github.com/IBM/storage-fusion.git|repoURL: https://github.com/<your-org>/storage-fusion.git|g'

# Update branch/tag
find AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu \
  -name '*.yaml' | xargs sed -i '' \
  's|targetRevision: master|targetRevision: <your-branch>|g'
```

If the `appproject-*.yaml` files contain an explicit `sourceRepos` allowlist (not a wildcard `*`), add your new `repoURL` there too. ArgoCD will reject syncs from unlisted repos:

```yaml
# appproject-prod.yaml
spec:
  sourceRepos:
    - https://github.com/IBM/storage-fusion.git            # existing
    - https://github.com/<your-org>/storage-fusion.git     # add your fork
```

**Commit and push your changes:**

```bash
git add AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu/
git commit -m "Update repoURL and targetRevision for fork"
git push
```

---

### Step 4: Deploy Model AppProjects and Applications

The AppProject is shared across all three model Applications and only needs to be applied once. Each model then gets its own Application manifest.

**Apply the shared AppProject:**

```bash
cd AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu

oc apply -f environments/prod/appproject-prod.yaml
```

**Verify the AppProject:**

```bash
oc get appproject fusion-vllm-cpu-model-deploy-prod -n openshift-gitops
```

**Register all three model Applications with ArgoCD:**

```bash
oc apply -f environments/prod/application-qwen2-5-1-5b-cpu.yaml
oc apply -f environments/prod/application-qwen2-5-coder-1-5b-cpu.yaml
oc apply -f environments/prod/application-smollm2-1-7b-cpu.yaml
```

Verify all three are registered:

```bash
oc get applications.argoproj.io -n openshift-gitops \
  -l component=model-deploy-cpu,environment=prod
```

**Expected output:**
```text
NAME                                                 SYNC STATUS   HEALTH STATUS
fusion-vllm-cpu-model-deploy-prod-qwen2-5-1-5b       OutOfSync     Missing
fusion-vllm-cpu-model-deploy-prod-qwen2-5-coder-1-5b OutOfSync     Missing
fusion-vllm-cpu-model-deploy-prod-smollm2-1-7b        OutOfSync     Missing
```

`OutOfSync / Missing` is expected because the resources do not exist on the cluster yet.

---

### Step 5: Sync the Model Applications

Production uses **manual sync** for change control. Each Application must be synced explicitly.

#### Default: ESO / Vault

ESO is enabled by default in the prod values files. All three prod models share a **single Vault secret** containing the S3 credentials. Before syncing for the first time, verify the shared secret exists at the correct path:

```bash
# Verify the shared secret exists and has all five required keys
vault kv get secret/maas/model-registry/object-storage
# Must show: access_key_id, secret_access_key, endpoint, region, bucket
```

If the secret does not exist, create it:

```bash
vault kv put secret/maas/model-registry/object-storage \
  access_key_id="<key>" \
  secret_access_key="<secret>" \
  endpoint="https://s3.openshift-storage.svc:443" \
  region="us-south" \
  bucket="<bucket-name>"
```

> All three models (`qwen2-5-1-5b-cpu`, `qwen2-5-coder-1-5b-cpu`, `smollm2-1-7b-cpu`) share
> the same S3 bucket and credentials. Their environment values files all point
> `s3.externalSecret.remoteRef.key` to this single path. **One write, all three models covered.**

Sync all three Applications. ESO reconciles the shared `ExternalSecret` CRs into the
`deploy-models-cpu-connection` and `storage-config` Secrets automatically:

```bash
argocd app sync fusion-vllm-cpu-model-deploy-prod-qwen2-5-1-5b
argocd app sync fusion-vllm-cpu-model-deploy-prod-qwen2-5-coder-1-5b
argocd app sync fusion-vllm-cpu-model-deploy-prod-smollm2-1-7b
```

After sync, confirm ESO reconciled successfully:

```bash
# ExternalSecrets should show SecretSynced: True
oc get externalsecret -n deploy-models-cpu \
  -o custom-columns='NAME:.metadata.name,READY:.status.conditions[0].status,REASON:.status.conditions[0].reason'

# Backing Secrets should exist with all AWS_* keys populated
oc get secret deploy-models-cpu-connection storage-config -n deploy-models-cpu
```

#### Alternative: Manual credentials at sync time

To disable ESO and pass credentials manually, set `s3.externalSecret.enabled: false` in each
environment values file, then sync with `--helm-set` flags:

```bash
argocd app sync fusion-vllm-cpu-model-deploy-prod-qwen2-5-1-5b \
  --helm-set s3.externalSecret.enabled=false \
  --helm-set s3.accessKeyId=<your-access-key> \
  --helm-set s3.secretAccessKey=<your-secret-key> \
  --helm-set s3.bucket=<your-bucket> \
  --helm-set s3.region=us-south

argocd app sync fusion-vllm-cpu-model-deploy-prod-qwen2-5-coder-1-5b \
  --helm-set s3.externalSecret.enabled=false \
  --helm-set s3.accessKeyId=<your-access-key> \
  --helm-set s3.secretAccessKey=<your-secret-key> \
  --helm-set s3.bucket=<your-bucket> \
  --helm-set s3.region=us-south

argocd app sync fusion-vllm-cpu-model-deploy-prod-smollm2-1-7b \
  --helm-set s3.externalSecret.enabled=false \
  --helm-set s3.accessKeyId=<your-access-key> \
  --helm-set s3.secretAccessKey=<your-secret-key> \
  --helm-set s3.bucket=<your-bucket> \
  --helm-set s3.region=us-south
```

#### Order of first sync: shared resources

The first Application to sync creates the shared Namespace, ServiceAccount, `ExternalSecret` CRs
(`deploy-models-cpu-connection`, `storage-config`), and `argocd-manager-vllm-cpu`
ClusterRole/RoleBinding. All carry `helm.sh/resource-policy: keep`. The second and third
Applications see those resources already exist and skip creation without error. ArgoCD suppresses
label-drift on shared resources via the `ignoreDifferences` blocks in each Application manifest.

> **Sync-wave ordering** ensures the `InferenceService` is only applied after ESO has
> reconciled the `ExternalSecret` CRs into real Secrets (wave 0 → RBAC, wave 1 → ExternalSecrets,
> wave 2 → InferenceService). This prevents the RHOAI admission webhook from rejecting the
> `InferenceService` CREATE because the connection Secret does not yet exist.

**Check sync and health status after syncing:**

```bash
oc get applications.argoproj.io -n openshift-gitops \
  -l component=model-deploy-cpu,environment=prod \
  -o custom-columns='NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
```

**Expected output:**
```text
NAME                                                   SYNC     HEALTH        REVISION
fusion-vllm-cpu-model-deploy-prod-qwen2-5-1-5b         Synced   Progressing   <commit-sha>
fusion-vllm-cpu-model-deploy-prod-qwen2-5-coder-1-5b   Synced   Progressing   <commit-sha>
fusion-vllm-cpu-model-deploy-prod-smollm2-1-7b         Synced   Progressing   <commit-sha>
```

`Progressing` is expected because pods are starting up and downloading model weights from S3.

---

### Step 6: Monitor Model Pods

CPU models require time to:
1. Pull the vLLM CPU container image
2. Download model artifacts from ODF storage (storage-initializer `Init` container)
3. Run vLLM AOT compilation warmup (can take 5–10 minutes)

**Watch pod lifecycle:**

```bash
# Stream pod status changes
oc get pods -n deploy-models-cpu -w

# Follow storage-initializer logs (model download progress)
oc logs -n deploy-models-cpu \
  $(oc get pod -n deploy-models-cpu \
    -l serving.kserve.io/inferenceservice=qwen2-5-1-5b-cpu \
    -o jsonpath='{.items[0].metadata.name}') \
  -c storage-initializer --follow

# Follow vLLM startup logs (after Init completes)
oc logs -n deploy-models-cpu \
  -l serving.kserve.io/inferenceservice=qwen2-5-1-5b-cpu \
  -c kserve-container -f
```

**Watch all three InferenceServices become Ready:**

```bash
oc get inferenceservice -n deploy-models-cpu -w
```

**Expected output (after all pods are ready):**
```text
NAME                     URL                                                                              READY   AGE
qwen2-5-1-5b-cpu         https://qwen2-5-1-5b-cpu-deploy-models-cpu.apps.<cluster-domain>                True    12m
qwen2-5-coder-1-5b-cpu   https://qwen2-5-coder-1-5b-cpu-deploy-models-cpu.apps.<cluster-domain>          True    13m
smollm2-1-7b-cpu         https://smollm2-1-7b-cpu-deploy-models-cpu.apps.<cluster-domain>                True    14m
```

If any model shows `READY=False` after 15 minutes, see [Troubleshooting](#troubleshooting).

**Check events for any issues:**

```bash
oc get events -n deploy-models-cpu --sort-by='.lastTimestamp' | tail -20
```

---

### Step 7: Test Inference Endpoints

Once the InferenceServices report `READY=True`, validate each model with live API calls.

#### Set up tokens and endpoints

Each model has its own bearer-token Secret and Route. Export all variables in one block:

```bash
# Tokens
TOKEN=$(oc get secret external-access-token-qwen2-5-1-5b-cpu-sa \
  -n deploy-models-cpu -o jsonpath='{.data.token}' | base64 -d)
TOKEN_CODER=$(oc get secret external-access-token-qwen2-5-coder-1-5b-cpu-sa \
  -n deploy-models-cpu -o jsonpath='{.data.token}' | base64 -d)
TOKEN_SMOL=$(oc get secret external-access-token-smollm2-1-7b-cpu-sa \
  -n deploy-models-cpu -o jsonpath='{.data.token}' | base64 -d)

# Endpoints
ENDPOINT_QWEN=$(oc get inferenceservice qwen2-5-1-5b-cpu \
  -n deploy-models-cpu -o jsonpath='{.status.url}')
ENDPOINT_CODER=$(oc get inferenceservice qwen2-5-coder-1-5b-cpu \
  -n deploy-models-cpu -o jsonpath='{.status.url}')
ENDPOINT_SMOL=$(oc get inferenceservice smollm2-1-7b-cpu \
  -n deploy-models-cpu -o jsonpath='{.status.url}')
```

#### Test all three models

```bash
# Qwen2.5-1.5B: general instruction following
curl -sk -X POST "${ENDPOINT_QWEN}/v1/chat/completions" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"model":"qwen2-5-1-5b-cpu","messages":[{"role":"user","content":"What is IBM Fusion and how does it support AI workloads?"}],"max_tokens":150}'

# Qwen2.5-Coder-1.5B: code generation
curl -sk -X POST "${ENDPOINT_CODER}/v1/chat/completions" \
  -H "Authorization: Bearer ${TOKEN_CODER}" \
  -H "Content-Type: application/json" \
  -d '{"model":"qwen2-5-coder-1-5b-cpu","messages":[{"role":"user","content":"Write a Python function that checks if a string is a palindrome."}],"max_tokens":200}'

# SmolLM2-1.7B: compact general purpose
curl -sk -X POST "${ENDPOINT_SMOL}/v1/chat/completions" \
  -H "Authorization: Bearer ${TOKEN_SMOL}" \
  -H "Content-Type: application/json" \
  -d '{"model":"smollm2-1-7b-cpu","messages":[{"role":"user","content":"Summarise the benefits of CPU-based LLM inference in two sentences."}],"max_tokens":150}'
```

Each response follows the OpenAI `chat.completion` schema:

```json
{
    "id": "chatcmpl-...",
    "object": "chat.completion",
    "model": "<model-name>",
    "choices": [{"message": {"role": "assistant", "content": "..."}, "finish_reason": "stop"}],
    "usage": {"prompt_tokens": 33, "completion_tokens": 150, "total_tokens": 183}
}
```

---

## What's Deployed

**Shared resources** (created once, retained across model syncs via `resource-policy: keep`):

| Component | Kind | Notes |
|---|---|---|
| `fusion-vllm-cpu-model-deploy-prod` | `AppProject` | Shared by all prod model Apps |
| `deploy-models-cpu` | `Namespace` | Shared namespace |
| `argocd-manager-vllm-cpu` | `ClusterRole` + `RoleBinding` | KServe + ESO RBAC for ArgoCD |
| `deploy-models-cpu-connection-sa` | `ServiceAccount` | Shared S3 storage SA |
| `deploy-models-cpu-connection` | `Secret` | Shared ODF S3 credentials |
| `storage-config` | `Secret` | KServe storage-initializer config |

**Per-model resources** (one set per deployed model):

| Component | Kind | Notes |
|---|---|---|
| `<model>` | `ServingRuntime` | vLLM CPU runtime |
| `<model>` | `InferenceService` | KServe v1beta1 |
| `<model>-sa` | `ServiceAccount` | Bearer-token SA for external access |
| `external-access-token-<model>-sa` | `Secret` | SA token, auto-populated by OpenShift |
| `deploy-models-cpu-<model>-sa-auth-delegator` | `ClusterRoleBinding` | Required for `kube-rbac-proxy` TokenReview |
| `<model>-sa-inferenceservice-viewer` | `Role` + `RoleBinding` | Grants SAR check for `kube-rbac-proxy` |

---

## Add a New CPU Model

To deploy a fourth CPU model alongside the three already running:

**1. Create a values file for the new model:**

```bash
cp AI/quickstarts/model-as-a-service/deploy/helm/model-deploy-vllm-cpu/environments/prod/values-qwen2-5-1-5b-cpu.yaml \
   AI/quickstarts/model-as-a-service/deploy/helm/model-deploy-vllm-cpu/environments/prod/values-<model-name>.yaml
```

Edit the new file and update these fields:
- `model.name`: kebab-case Kubernetes name (e.g. `gemma-2-2b-cpu`)
- `model.displayName`: shown in RHOAI dashboard
- `s3.modelPath`: path to the model artifacts within the ODF bucket
- `resources`: size for the model's parameter count (see table below)

| Model size | dtype | Recommended limits |
|---|---|---|
| ≤ 1.5B | `bfloat16` | `cpu: 4, memory: 12Gi` |
| ≤ 2B | `bfloat16` | `cpu: 4, memory: 14Gi` |
| 3B | `bfloat16` | `cpu: 8, memory: 16Gi` |
| 7B | `bfloat16` | `cpu: 16, memory: 32Gi` |

**2. Create the ArgoCD Application manifest:**

```bash
cp AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu/environments/prod/application-qwen2-5-1-5b-cpu.yaml \
   AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu/environments/prod/application-<model-name>.yaml
```

Edit the new Application manifest and update:
- `metadata.name`: unique ArgoCD Application name
- `metadata.labels.model`
- `helm.valueFiles`: point to your new values file
- `info`: description fields

**3. Commit, push, and deploy:**

```bash
git add AI/quickstarts/model-as-a-service/deploy/helm/model-deploy-vllm-cpu/environments/prod/values-<model-name>.yaml
git add AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu/environments/prod/application-<model-name>.yaml
git commit -m "Add <model-name> CPU model"
git push

# The AppProject is already applied. Just register and sync the new Application.
oc apply -f AI/quickstarts/model-as-a-service/deploy/gitops/model-deploy-vllm-cpu/environments/prod/application-<model-name>.yaml
argocd app sync fusion-vllm-cpu-model-deploy-prod-<model-name>
```

> The new Application will create its own `ServingRuntime` and `InferenceService` but reuse the existing shared `Namespace`, `ServiceAccount`, `storage-config`, and `ClusterRole`. The `ignoreDifferences` block in the new Application manifest prevents OutOfSync on those shared resources.

---

## Troubleshooting

### Pod stuck in `Init` (model download in progress)

```bash
# Check what the storage-initializer is doing
oc logs -n deploy-models-cpu \
  $(oc get pod -n deploy-models-cpu \
    -l serving.kserve.io/inferenceservice=qwen2-5-1-5b-cpu \
    -o jsonpath='{.items[0].metadata.name}') \
  -c storage-initializer --follow
```

This is normal. `Init` means the model artifacts are downloading from ODF storage. Wait 3–5 minutes for a ~3 GiB model.

### `Storage initialization failed: No model found`

The `s3.modelPath` value does not match the actual key prefix in the ODF bucket. List the bucket contents to find the correct path:

```bash
# Get credentials from the ESO-managed Secret
ACCESS_KEY=$(oc get secret deploy-models-cpu-connection -n deploy-models-cpu \
  -o jsonpath='{.data.AWS_ACCESS_KEY_ID}' | base64 -d)
SECRET_KEY=$(oc get secret deploy-models-cpu-connection -n deploy-models-cpu \
  -o jsonpath='{.data.AWS_SECRET_ACCESS_KEY}' | base64 -d)
BUCKET=$(oc get secret deploy-models-cpu-connection -n deploy-models-cpu \
  -o jsonpath='{.data.AWS_S3_BUCKET}' | base64 -d)

kubectl run s3-ls --rm -i --restart=Never --image=amazon/aws-cli -n deploy-models-cpu \
  --env="AWS_ACCESS_KEY_ID=${ACCESS_KEY}" \
  --env="AWS_SECRET_ACCESS_KEY=${SECRET_KEY}" \
  --env="AWS_DEFAULT_REGION=us-south" \
  --command -- aws s3 ls s3://${BUCKET}/ \
  --endpoint-url https://s3.openshift-storage.svc:443 --no-verify-ssl
```

Update `s3.modelPath` in the environment values file to match the exact prefix shown (keys use **hyphens**, not dots).

### Pod `OOMKilled` (exit code 137)

The container exceeded its memory limit during vLLM AOT compilation warmup. Common causes:

1. **`--dtype=float32` in `extraArgs`**: doubles weight memory vs `bfloat16` (~5.6 GiB vs ~3 GiB). Remove it.
2. **`--max-model-len` set**: paradoxically inflates KV cache slot count (112,256 tokens with 4096 ctx), consuming more RAM. Remove it.
3. **`VLLM_CPU_KVCACHE_SPACE` too large**: reduce to `container_limit - weights_bfloat16 - 2 GiB`.

The working formula: `VLLM_CPU_KVCACHE_SPACE = memory_limit - 3 GiB (weights) - 2 GiB (overhead)`. For the 12 GiB prod limit this gives `6`.

### External route returns `403 Forbidden` with a valid SA token

The `kube-rbac-proxy` sidecar (injected by RHOAI into every `InferenceService` pod) performs two checks:
1. **TokenReview**: validates the bearer JWT. Requires `system:auth-delegator` at cluster scope.
2. **SubjectAccessReview**: checks the token identity has `get` on the specific `InferenceService`.

Both RBAC bindings are created automatically by the chart (`templates/external-access-sa.yaml`) when `externalAccess.enabled: true`. Verify they exist:

```bash
# Check ClusterRoleBinding for TokenReview
oc get clusterrolebinding \
  deploy-models-cpu-qwen2-5-1-5b-cpu-sa-auth-delegator

# Check Role + RoleBinding for SAR
oc get role,rolebinding \
  qwen2-5-1-5b-cpu-sa-inferenceservice-viewer \
  -n deploy-models-cpu
```

If missing, re-sync the Application to create them.

### Pod stuck in `Pending` (insufficient resources)

```bash
oc describe pod -n deploy-models-cpu \
  -l serving.kserve.io/inferenceservice=qwen2-5-1-5b-cpu
```

Look for `Insufficient cpu` or `Insufficient memory` in the Events section. Either increase `resources.requests` to schedulable values or ensure an appropriate node exists.

### `ImagePullBackOff` on the vLLM CPU image

The registry `registry.redhat.io/rhaii-early-access/vllm-cpu-rhel9` does not support `:latest`. The values files pin to a digest. Verify the image value is a full `@sha256:...` digest, not a tag:

```yaml
servingRuntime:
  image: "registry.redhat.io/rhaii-early-access/vllm-cpu-rhel9@sha256:139e070c2e4e2bc0254aa852b436c35ec2659e8424e228cc4c77a3bc2ddf50b6"
```

### Application OutOfSync on Namespace / Secret / ServiceAccount / ClusterRole

`ServerSideApply` stamps instance-specific labels on shared resources on every sync. All Application manifests include `ignoreDifferences` entries to suppress this. If you still see OutOfSync:

```bash
# Confirm RespectIgnoreDifferences is enabled
oc get application fusion-vllm-cpu-model-deploy-prod-qwen2-5-1-5b \
  -n openshift-gitops \
  -o jsonpath='{.spec.syncPolicy.syncOptions}'
```

The output must include `RespectIgnoreDifferences=true`. If missing, add it to the Application's `syncPolicy.syncOptions`.

### Connection Secret not found / `storage-config` missing

In ESO mode (the prod default), the `deploy-models-cpu-connection` and `storage-config` Secrets
are created by ESO reconciling the `ExternalSecret` CRs, not by an ODF ObjectBucketClaim.
If these Secrets are missing after sync, check the `ExternalSecret` status first:

```bash
# Check ExternalSecret reconciliation status
oc get externalsecret -n deploy-models-cpu \
  -o custom-columns='NAME:.metadata.name,READY:.status.conditions[0].status,REASON:.status.conditions[0].reason,MESSAGE:.status.conditions[0].message'

# If SyncFailed, describe for the full error
oc describe externalsecret deploy-models-cpu-connection -n deploy-models-cpu
oc describe externalsecret storage-config -n deploy-models-cpu
```

Common causes:

| Symptom | Cause | Fix |
|---|---|---|
| `SecretSyncedError: secret not found` | Vault path `maas/model-registry/object-storage` does not exist | Run `vault kv put` (see Step 5) |
| `SecretSyncedError: store not ready` | `ClusterSecretStore/vault-backend` is not `Ready` | Run `oc get clustersecretstore vault-backend` and check the status |
| `ExternalSecret` objects not created at all | `InferenceService` webhook blocked the sync before wave 1 | Check `oc get application ... -o jsonpath='{.status.operationState.message}'` |

Verify the Vault secret exists with all five required keys:

```bash
vault kv get secret/maas/model-registry/object-storage
# Must show: access_key_id, secret_access_key, endpoint, region, bucket
```

---

## Next Steps

After completing this quickstart, you can:

1. **Scale to GPU with MaaS**: when GPU nodes become available, the same GitOps pattern extends to `LLMInferenceService` with the full MaaS gateway, subscriptions, and API-key enforcement. See [Quickstart: IBM Fusion Model-as-a-Service Platform (GitOps Deployment)](https://github.com/IBM/storage-fusion/blob/master/AI/quickstarts/model-as-a-service/infoDocs/gitops-deployment-guide_updated.md)
2. **Build a multi-model chat application**: the three deployed models are immediately usable from any OpenAI-SDK-compatible client. See the [Agentic Chat Assistant Sample Application](https://github.com/IBM/storage-fusion/blob/master/AI/fusion-gitops-sample-app/README.md) for a reference implementation

---

## Summary

This quickstart demonstrated how to deploy CPU-based LLM inference services on IBM Fusion using GitOps. You now have:

- **OpenShift AI (RHOAI) as the model platform**: KServe model serving, Model Registry, and `ServingRuntime`/`InferenceService` lifecycle managed by RHOAI
- **Three CPU LLMs deployed**: Qwen2.5-1.5B, Qwen2.5-Coder-1.5B, SmolLM2-1.7B
- **GitOps-managed**: all resources declared in Git, reconciled by ArgoCD
- **Shared namespace design**: Namespace, ServiceAccount, connection Secret, and storage-config shared across models with `helm.sh/resource-policy: keep`
- **Per-model ServingRuntime**: isolated vLLM CPU instances, no runtime conflicts
- **External access with bearer-token auth**: OpenShift Routes with TLS reencrypt
- **ODF object storage**: model artifacts auto-provisioned via ObjectBucketClaim, no manual credential setup
- **Validated inference endpoints**: chat completions API tested end-to-end

The IBM Fusion-hosted CPU inference platform is now operational for development, validation, and functional testing. Models are accessible through secured, GitOps-managed OpenShift Routes and ready to serve as backends for evaluation and hands-on learning. When you are ready to scale to production-grade workloads, see [Next Steps](#next-steps) for the GPU/MaaS upgrade path.
