// internal/server/context.go
package server

import (
	"sync"
	"time"
)

type Context struct {
	mu     sync.RWMutex
	repo   string
	token  string
	expiry time.Time
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) SetRepo(repo string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.repo = repo
}

func (c *Context) GetRepo() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repo
}

func (c *Context) SetToken(token string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	c.expiry = time.Now().Add(ttl)
}

func (c *Context) GetToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.token == "" || time.Now().After(c.expiry) {
		return "" // expired or not set
	}
	return c.token
}
