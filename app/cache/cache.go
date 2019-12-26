// Package cache is the wrapper around patrickmn/go-cache package with different delete modes.
package cache

import (
	"fmt"
	"regexp"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Delete mode possible values
const (
	DeleteExpire         = "expire"
	DeleteComplete       = "complete"
	DeleteExpireComplete = "expire_complete"
)

// Cache contains go-cache.Cache and deleteMode.
type Cache struct {
	c          *gocache.Cache
	deleteMode string
	excludes   []*regexp.Regexp
}

// New creates new Cache.
func New(expire time.Duration, deleteMode string, excludes []*regexp.Regexp) (*Cache, error) {
	if !(deleteMode == DeleteExpire || deleteMode == DeleteComplete || deleteMode == DeleteExpireComplete) {
		return nil, fmt.Errorf("wrong cache delete mode: %v", deleteMode)
	}

	c := gocache.New(expire, 2*expire)
	return &Cache{
		c:          c,
		deleteMode: deleteMode,
		excludes:   excludes,
	}, nil
}

// Contains returns true if key `k` was found in cache.
func (c *Cache) Contains(k string) bool {
	_, found := c.c.Get(k)
	return found
}

// Set add item with key `k` and value `v` in cache based on deleteMode.
func (c *Cache) Set(k string, v interface{}) {
	for _, r := range c.excludes {
		if r.MatchString(k) {
			return
		}
	}
	switch c.deleteMode {
	case DeleteExpire, DeleteExpireComplete:
		c.c.Set(k, v, gocache.DefaultExpiration)
	case DeleteComplete:
		c.c.Set(k, v, gocache.NoExpiration)
	}
}

// Del deletes items from cache based on deleteMode.
func (c *Cache) Del(k string) {
	switch c.deleteMode {
	case DeleteComplete, DeleteExpireComplete:
		c.c.Delete(k)
	}
}
