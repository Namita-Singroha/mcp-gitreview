# MCP GitReview Server

🚀 A lightweight [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for exploring GitHub repositories directly through LLMs.  

This project lets an AI agent interact with GitHub repos in a structured way — listing commits, branches, PRs, and repo metadata — without exposing raw `git` or GitHub API complexity.  

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

### 3. Set a repo
```json
{
  "tool": "set_repo",
  "arguments": { "repo": "modelcontextprotocol/go-sdk" }
}
```

### 4. (Optional) Authenticate for private repos
```json
{
  "tool": "set_github_token",
  "arguments": { "token": "ghp_XXXXXXXXXXXXXXXX" }
}
```

---

## 🤖 Example Flow

1. Ask: *“List the branches in `modelcontextprotocol/go-sdk`”*  
2. LLM realizes no repo is set → calls `set_repo`  
3. Server stores it → responds: *“Repo set to modelcontextprotocol/go-sdk”*  
4. LLM calls `list_branches` → gets actual branch names.  
