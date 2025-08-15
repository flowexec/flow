package cache

func UpdateAll() error {
	wsCache := NewWorkspaceCache()
	if err := wsCache.Update(); err != nil {
		return err
	}

	execCache := NewExecutableCache(wsCache)
	if err := execCache.Update(); err != nil {
		return err
	}

	return nil
}
