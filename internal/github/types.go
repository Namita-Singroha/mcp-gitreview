package github

// Commit represents a commit from GitHub API
type Commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

// Branch represents a branch from GitHub API
type Branch struct {
	Name string `json:"name"`
}