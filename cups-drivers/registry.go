package drivers

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = map[string]Driver{}
)

// Register adds a driver to the global registry.
// Panics if id is empty or already registered.
func Register(d Driver) {
	if d == nil {
		panic("cups-drivers: Register(nil)")
	}
	id := d.ID()
	if id == "" {
		panic("cups-drivers: driver ID must not be empty")
	}

	mu.Lock()
	defer mu.Unlock()
	if _, exists := registry[id]; exists {
		panic(fmt.Sprintf("cups-drivers: driver %q already registered", id))
	}
	registry[id] = d
}

// Get returns a registered driver by ID.
func Get(id string) (Driver, error) {
	mu.RLock()
	defer mu.RUnlock()
	d, ok := registry[id]
	if !ok {
		return nil, fmt.Errorf("cups-drivers: unknown model %q (available: %v)", id, listLocked())
	}
	return d, nil
}

// MustGet returns a registered driver or panics.
func MustGet(id string) Driver {
	d, err := Get(id)
	if err != nil {
		panic(err)
	}
	return d
}

// List returns sorted driver IDs.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	return listLocked()
}

func listLocked() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
