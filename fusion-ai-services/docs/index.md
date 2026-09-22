# IBM Fusion AI Services — Getting Started

**IBM Fusion** is an enterprise data and AI platform built on Red Hat OpenShift —
on-prem, private cloud, or at the edge. It ships storage, data governance, and
AI-ready services as native OpenShift capabilities. Your data stays on your
infrastructure. Your AI workloads reach it instantly.

This Developer Hub gives every developer on your team a single place to discover
what Fusion AI Services are running, access their endpoints, and connect AI tools —
without needing cluster access or tracking down URLs.

---

## Services in this Developer Hub

These are full enterprise services — not thin API wrappers. MCP is an additional
integration layer that CAS and DCS expose for AI tooling.

| Service | What it is | What you can do with it |
|---|---|---|
| **CAS** — Content Aware Storage | Intelligent vector storage on Fusion | Store and semantically search documents, images, and unstructured data at scale |
| **DCS** — Data Cataloging Service | Data governance and discovery engine | Discover, tag, classify, and govern data assets across your cluster |
| **WXO** — watsonx Orchestrate | AI agent platform via Fusion Software Hub | Build and run AI agents on OpenShift — wire CAS and DCS into agent toolsets |

---

## Step 1 — Check what's already registered

Before doing anything, check whether your service is already in the catalog:

👉 [Browse registered Fusion AI Services](/catalog?filters%5Bkind%5D=component&filters%5Btype%5D=fusion-service&limit=20)

If your cluster is already listed:

1. Click the component
2. Go to the **Links** tab
3. Console, API docs, health check, and MCP endpoints are all pre-filled — no cluster login needed

**Already there? Jump to [Step 3](#step-3--use-it).**

---

## Step 2 — Register your service

Not listed? Use one of the self-service templates below. You only need two things:
a short cluster name and the OCP API URL (`https://api.<domain>:6443`).
Every endpoint is derived automatically — no token, no OCP login required.

| I have... | Template to use |
|---|---|
| CAS installed on my OpenShift cluster | [Add IBM Fusion CAS Cluster](/create/templates/default/add-cas-cluster) |
| DCS (Spectrum Discover) on my cluster | [Add IBM Fusion DCS Cluster](/create/templates/default/add-dcs-cluster) |
| watsonx Orchestrate deployed via Software Hub | [Register watsonx Orchestrate Instance](/create/templates/default/add-wxo-instance) |

After the template runs, your service appears in the catalog with all endpoints
pre-filled in the **Links** tab.

---

## Step 3 — Use it

Open the registered catalog component → **Links** tab.

Everything is pre-filled:

- **Console** — open the service UI directly, no cluster login needed
- **Swagger / API docs** — explore the REST API
- **Health check** — verify the service is up
- **MCP endpoints** (CAS and DCS only) — copy and paste into your AI tool

To connect an AI tool via MCP, copy the endpoint from the Links tab and follow
the documentation for your specific tool:

- [CAS MCP Integration — IBM Docs](https://www.ibm.com/docs/en/fusion-software/2.13.0?topic=cas-integrating-model-context-protocol-mcp)
- [DCS MCP Server — IBM Docs](https://www.ibm.com/docs/en/fusion-software/2.13.0?topic=capabilities-data-cataloging-mcp-server)
- [CAS + watsonx Orchestrate via MCP — IBM Docs](https://www.ibm.com/docs/en/fusion-software/2.13.0?topic=cas-integrating-watsonx-orchestrate-by-using-mcp)

---

## MCP Tools Reference

CAS and DCS each expose an MCP Server alongside their main APIs.
These are the operations your AI agent can call once connected.

### CAS — Content Aware Storage

| Tool | What it does |
|---|---|
| `list_vector_stores` | List all vector stores on this CAS instance |
| `search_vector_stores` | Semantic search across stored documents, files, and images |
| `get_vector_store_file_content` | Retrieve file content from a search result |

📖 [CAS MCP Integration](https://www.ibm.com/docs/en/fusion-software/2.13.0?topic=cas-integrating-model-context-protocol-mcp) · [CAS + watsonx Orchestrate](https://www.ibm.com/docs/en/fusion-software/2.13.0?topic=cas-integrating-watsonx-orchestrate-by-using-mcp) · [Blog: Bob meets CAS MCP](https://community.ibm.com/community/user/blogs/namita-singroha/2026/03/24/bob-meets-ibm-cas-mcp)

### DCS — Data Cataloging Service

| Tool | What it does |
|---|---|
| `dcs_file_search` | Search data assets in the catalog |
| `dcs_get_registered_tags` | List all metadata tags |
| `dcs_get_recommend_tags` | AI-suggested tags for a dataset |
| `dcs_create_tag` | Create a new metadata tag |
| `dcs_create_policy` | Define a data governance policy |
| `dcs_get_policies` | List all active policies |
| `dcs_set_credentials` | Configure credentials for a data source |

📖 [DCS MCP Server](https://www.ibm.com/docs/en/fusion-software/2.13.0?topic=capabilities-data-cataloging-mcp-server) · [Blog: Explore catalog with NLP](https://community.ibm.com/community/user/blogs/paul-llamas-virgen/2026/03/06/exploring-your-fusion-data-catalog-with-nlp) · [Blog: AI Agents on WXO for DCS](https://community.ibm.com/community/user/blogs/paul-llamas-virgen/2026/03/06/create-your-ai-agents-on-wx-orch-for-data-catalog)

### WXO — watsonx Orchestrate

After registering, the catalog entry links directly to your WXO UI, ADK documentation,
and IBM Docs. From there you can build agents and wire CAS or DCS as MCP-based tools.

📖 [watsonx Orchestrate Docs](https://www.ibm.com/docs/en/watsonx/watson-orchestrate) · [ADK: Build an Agent](https://developer.watson-orchestrate.ibm.com/ai_builder/creating_agent) · [Blog: AI Agents on WXO for DCS](https://community.ibm.com/community/user/blogs/paul-llamas-virgen/2026/03/06/create-your-ai-agents-on-wx-orch-for-data-catalog)

---

## IBM Documentation

- [IBM Fusion Knowledge Center](https://www.ibm.com/docs/en/fusion-software)
- [IBM Fusion HCI Knowledge Center](https://www.ibm.com/docs/en/fusion-hci-systems)
- [IBM Fusion Tech Community](https://ibm.github.io/storage-fusion/fusion-ai/overview/)
- [IBM Tech Exchange](https://community.ibm.com/community/user/groups/community-home/recent-community-blogs?communitykey=e596ba82-cd57-4fae-8042-163e59279ff3)
