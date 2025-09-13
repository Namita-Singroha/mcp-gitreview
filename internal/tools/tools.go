package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Namita-Singroha/mcp-gitreview/internal/github"
	"github.com/Namita-Singroha/mcp-gitreview/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all tools with the MCP server
func RegisterAll(mcpServer *mcp.Server, serverCtx *server.Context) {
	registerSetRepoTool(mcpServer, serverCtx)
	registerSetTokenTool(mcpServer, serverCtx)
	registerListCommitsTool(mcpServer, serverCtx)
	registerListBranchesTool(mcpServer, serverCtx)
	registerGetLatestCommitTool(mcpServer, serverCtx)
	registerGetRepoInfoTool(mcpServer, serverCtx)
	registerListPullRequestsTool(mcpServer, serverCtx)
}

// --- Helpers ---
func getRepoFromArgs(args any, serverCtx *server.Context) string {
	var input struct {
		Repo string `json:"repo"`
	}
	if args != nil {
		b, _ := json.Marshal(args)
		_ = json.Unmarshal(b, &input)
	}
	if input.Repo != "" {
		return input.Repo
	}
	return serverCtx.GetRepo()
}

func newGitClient(serverCtx *server.Context) *github.Client {
	return github.NewClientFunc(func() string { return serverCtx.GetToken() })
}

func requireRepo(serverCtx *server.Context, args any) (string, *mcp.CallToolResult) {
	repo := getRepoFromArgs(args, serverCtx)
	if repo == "" {
		return "", server.ErrorResponse("No repository set. Please call `set_github_reposiory` first.")
	}
	return repo, nil
}

// --- Tool registrations ---

func registerSetRepoTool(mcpServer *mcp.Server, serverCtx *server.Context) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "set_github_repository",
		Description: "Set GitHub repository for subsequent tools",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Repo string `json:"repo"`
	}) (*mcp.CallToolResult, any, error) {
		if args.Repo == "" {
			return server.ErrorResponse("Please provide a repo in 'owner/name' format"), nil, nil
		}
		serverCtx.SetRepo(args.Repo)
		return server.SuccessResponse(fmt.Sprintf("Repo set to %s", args.Repo)), nil, nil
	})
}

func registerSetTokenTool(mcpServer *mcp.Server, serverCtx *server.Context) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "set_github_token",
		Description: "Set GitHub Personal Access Token for private repos",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Token string `json:"token"`
	}) (*mcp.CallToolResult, any, error) {
		if args.Token == "" {
			return server.ErrorResponse("Please provide a GitHub token"), nil, nil
		}
		serverCtx.SetToken(args.Token, time.Hour)
		return server.SuccessResponse("GitHub token set successfully (valid for 1 hour)"), nil, nil
	})
}

func registerListCommitsTool(mcpServer *mcp.Server, serverCtx *server.Context) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_commits",
		Description: "Fetch last 10 commits",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		repo, errResp := requireRepo(serverCtx, args)
		if errResp != nil {
			return errResp, nil, nil
		}
		client := newGitClient(serverCtx)
		commits, err := client.FetchCommits(repo)
		if err != nil {
			return server.ErrorResponse(fmt.Sprintf("Error fetching commits: %v", err)), nil, nil
		}
		return server.SuccessResponse(commits), nil, nil
	})
}

func registerListBranchesTool(mcpServer *mcp.Server, serverCtx *server.Context) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_branches",
		Description: "List all branches",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		repo, errResp := requireRepo(serverCtx, args)
		if errResp != nil {
			return errResp, nil, nil
		}
		client := newGitClient(serverCtx)
		branches, err := client.FetchBranches(repo)
		if err != nil {
			return server.ErrorResponse(fmt.Sprintf("Error fetching branches: %v", err)), nil, nil
		}
		return server.SuccessResponse(branches), nil, nil
	})
}

func registerGetLatestCommitTool(mcpServer *mcp.Server, serverCtx *server.Context) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get_latest_commit_message",
		Description: "Get latest commit message",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		repo, errResp := requireRepo(serverCtx, args)
		if errResp != nil {
			return errResp, nil, nil
		}
		client := newGitClient(serverCtx)
		msg, err := client.FetchLatestCommitMessage(repo)
		if err != nil {
			return server.ErrorResponse(fmt.Sprintf("Error fetching latest commit: %v", err)), nil, nil
		}
		return server.SuccessResponse(msg), nil, nil
	})
}

func registerGetRepoInfoTool(mcpServer *mcp.Server, serverCtx *server.Context) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "get_repository_details",
		Description: "Get repository details",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		repo, errResp := requireRepo(serverCtx, args)
		if errResp != nil {
			return errResp, nil, nil
		}
		client := newGitClient(serverCtx)
		info, err := client.FetchRepoInfo(repo)
		if err != nil {
			return server.ErrorResponse(fmt.Sprintf("Error fetching repo info: %v", err)), nil, nil
		}
		return server.SuccessResponse(info), nil, nil
	})
}

func registerListPullRequestsTool(mcpServer *mcp.Server, serverCtx *server.Context) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_pull_requests",
		Description: "List open pull requests",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		repo, errResp := requireRepo(serverCtx, args)
		if errResp != nil {
			return errResp, nil, nil
		}
		client := newGitClient(serverCtx)
		prs, err := client.FetchPullRequests(repo)
		if err != nil {
			return server.ErrorResponse(fmt.Sprintf("Error fetching PRs: %v", err)), nil, nil
		}
		return server.SuccessResponse(prs), nil, nil
	})
}
