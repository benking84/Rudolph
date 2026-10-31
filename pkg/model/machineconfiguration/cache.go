package machineconfiguration

import (
	"sync"
	"time"

	"github.com/airbnb/rudolph/pkg/clock"
)

func GetCache(timeProvider clock.TimeProvider) Cache {
	return &ConcreteCache{
		cacheDuration: time.Hour * 1,
		rules:         make(map[string]concreteCacheEntry),
		clock:         timeProvider,
	}
}

// Cache
type Cache interface {
	Has(key string) bool
	Get(key string) *MachineConfiguration
	Set(key string, config *MachineConfiguration) bool
}

type ConcreteCache struct {
	mu            sync.RWMutex
	clock         clock.TimeProvider
	rules         map[string]concreteCacheEntry
	cacheDuration time.Duration
}

type concreteCacheEntry struct {
	item    *MachineConfiguration
	expires time.Time
}

func (c *ConcreteCache) Has(key string) bool {
	c.mu.RLock()
	entry, ok := c.rules[key]
	c.mu.RUnlock()
	if !ok {
		return false
	}

	return entry.expires.After(c.clock.Now().UTC())
}

func (c *ConcreteCache) Get(key string) *MachineConfiguration {
	c.mu.RLock()
	entry, ok := c.rules[key]
	c.mu.RUnlock()
	if !ok {
		return nil
	}
	return entry.item
}

func (c *ConcreteCache) Set(key string, config *MachineConfiguration) bool {
	entry := concreteCacheEntry{
		item:    config,
		expires: c.clock.Now().UTC().Add(c.cacheDuration),
	}
	c.mu.Lock()
	c.rules[key] = entry
	c.mu.Unlock()

	return true
}

const (
	CacheKeyGlobal = "global"
)
