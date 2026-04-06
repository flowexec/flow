package cache

import (
	"github.com/flowexec/flow/pkg/store"
)

func UpdateAll() error {
	ds, err := store.NewDataStore(store.Path())
	if err != nil {
		return err
	}
	defer ds.Close()

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
