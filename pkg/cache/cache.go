package cache

import (
	"github.com/flowexec/flow/pkg/store"
)

// UpdateAll refreshes all caches using the provided DataStore.
// The caller is responsible for the DataStore lifecycle (closing, etc.).
func UpdateAll(ds store.DataStore) error {
	wsCache := NewWorkspaceCache(ds)
	if err := wsCache.Update(); err != nil {
		return err
	}

	execCache := NewExecutableCache(wsCache, ds)
	if err := execCache.Update(); err != nil {
		return err
	}

	return nil
}
