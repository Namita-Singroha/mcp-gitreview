# MCP GitReview Server

🚀 A lightweight Model Context Protocol (MCP) server for exploring GitHub repositories over stdio.  

This project enables AI agents (e.g., Claude Desktop) to interact with GitHub repos in a structured way — listing commits, branches, PRs, and repo metadata — without exposing raw git or GitHub API complexity.

---

## ✨ Features

- **Repository selection** → Set or switch repositories at runtime  
- **Authentication** → Add your GitHub Personal Access Token (PAT) for private repos  
- **GitHub data tools**:
  - View last 10 commits  
  - List all branches  
  - Get the latest commit message  
  - Fetch repo metadata (stars ⭐, forks 🍴, open issues 🐞)  
  - List open pull requests  

All these tools are exposed as MCP endpoints that an LLM can call directly.

---

## 📂 Project Structure

```
internal/
  github/   → GitHub API client
  server/   → Context handling (repo + token)
  tools/    → MCP tool definitions
main.go     → Entry point (starts the MCP server)
```

---

## 🔧 Usage

### 1. Build the server
```bash
go build -o mcp-gitreview .
```

### 2. Run via MCP
Example config (Claude Desktop / MCP client):

```json
{
  "mcpServers": {
    "gitreview": {
      "command": "/path/to/mcp-gitreview"
    }
  }
}
```

👉 No default repo is required. If a tool is called without a repo set, the server will politely ask the LLM to set one using `set_repo`.

📸 Demo: Claude Desktop interacting with the MCP server by invoking MCP Tools. 
The user asks in natural language, and the server communicates back in simple English, describing repository features clearly.

<img width="1269" height="766" alt="Screenshot 2025-09-13 at 7 34 59 PM" src="https://github.com/user-attachments/assets/7c7f46d6-f22d-4f73-864c-83c6d5be3583" />
<img width="1270" height="762" alt="Screenshot 2025-09-13 at 7 35 57 PM" src="https://github.com/user-attachments/assets/a7636462-6464-4b2c-ba8a-4ed7da1f9aef" />
<img width="1317" height="762" alt="Screenshot 2025-09-13 at 7 36 35 PM" src="https://github.com/user-attachments/assets/3e68c0f3-31d1-4624-a78c-35e8a83c81ed" />
<img width="1319" height="762" alt="Screenshot 2025-09-13 at 7 36 50 PM" src="https://github.com/user-attachments/assets/a68cc93a-3a16-4541-b610-a3a03a0ee853" />

---


## 🤖 Example Flow

1. Ask: *“List the branches in `modelcontextprotocol/go-sdk`”*  
2. LLM realizes no repo is set → calls `set_repo`  
3. Server stores it → responds: *“Repo set to modelcontextprotocol/go-sdk”*  
4. LLM calls `list_branches` → gets actual branch names.  
