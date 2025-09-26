import { useCallback, useMemo, useState } from "react";
import { useDebouncedValue } from "@mantine/hooks";
import type { EnrichedExecutable } from "../../types/executable";
import { FilterState, WorkspaceOption } from "./Filters.tsx";
import { useExecutables } from "../../hooks/useExecutable.ts";
import { useAppContext } from "../../hooks/useAppContext.tsx";

const ROOT_NAMESPACE_LABEL = 'Root namespace';

export function useExecutablesPage() {
  const { workspaces, selectedWorkspace } = useAppContext();
  const [filterState, setFilterState] = useState<FilterState>({
    search: '',
    tags: [],
    workspace: selectedWorkspace || '',
    namespace: '',
    verb: '',
    visibility: '',
    type: '',
  });

  const [debouncedSearch] = useDebouncedValue(filterState.search, 300);

  const cliFilters = useMemo(() => ({
    workspace: filterState.workspace || null,
    namespace: filterState.namespace === ROOT_NAMESPACE_LABEL ? '' : filterState.namespace || null,
    tag: filterState.tags || null,
    verb: filterState.verb || null,
  }), [filterState.workspace, filterState.namespace, filterState.tags, filterState.verb]);

  const {
    executables,
    isExecutablesLoading: isLoading,
    executablesError: error,
    refreshExecutables,
  } = useExecutables(
    cliFilters.workspace,
    cliFilters.namespace,
    cliFilters.tag,
    cliFilters.verb,
    // For now, don't pass search to CLI - do it client-side for responsiveness
    null,
  );

  const filteredExecutables = useMemo(() => {
    let filtered = executables;

    const searchTerm = debouncedSearch.trim().toLowerCase();
    if (searchTerm) {
      filtered = filtered.filter((executable : EnrichedExecutable) =>
        executable.ref.toLowerCase().includes(searchTerm) ||
        executable.fullDescription?.toLowerCase().includes(searchTerm)
      );
    }

    if (filterState.visibility) {
      filtered = filtered.filter((e) : boolean => {
        let v = e.visibility;
        if (!v) {
          v = 'private'; // default
        }
        return v === filterState.visibility
      });
    }

    if (filterState.type) {
      const t = filterState.type;
      filtered = filtered.filter((e) : boolean => {
        if (t === 'command' && e.exec) return true;
        if (t === 'serial' && e.serial) return true;
        if (t === 'parallel' && e.parallel) return true;
        if (t === 'launch' && e.launch) return true;
        if (t === 'request' && e.request) return true;
        if (t === 'render' && e.render) return true;
        return false;
      });
    }

    // TODO: support filtering root namespace on the client side
    if (filterState.namespace === ROOT_NAMESPACE_LABEL) {
      filtered = filtered.filter(e => !e.namespace);
    }

    return filtered.sort((a : EnrichedExecutable, b : EnrichedExecutable) => a.ref.localeCompare(b.ref));
  }, [executables, debouncedSearch, filterState.namespace, filterState.tags, filterState.visibility, filterState.type]);

  // Filter Dropdown Options
  const workspaceOptions : WorkspaceOption[] = useMemo(() => {
    return Array.from(workspaces).map(ws => ({ value: ws.name, label: ws.displayName } as WorkspaceOption));
  }, [workspaces]);

  const namespaceOptions: string[] = useMemo(() => {
    let filtered: EnrichedExecutable[] = executables;
    if (filterState.workspace) {
      filtered = filtered.filter(e => e.workspace === filterState.workspace);
    }

    const namespaces = new Set<string>();
    let hasRootNamespace = false;

    filtered.forEach(e => {
      if (e.namespace) {
        namespaces.add(e.namespace);
      } else {
        hasRootNamespace = true;
      }
    });

    const options = Array.from(namespaces).sort();
    if (hasRootNamespace) {
      options.unshift(ROOT_NAMESPACE_LABEL);
    }

    return options;
  }, [executables, filterState.workspace]);

  const tagOptions = useMemo(() => {
    const tags = new Set<string>();
    executables.forEach(e => e.tags?.forEach(tag => tags.add(tag)));
    return Array.from(tags).sort();
  }, [executables, filterState.workspace, filterState.namespace]);

  const verbOptions = useMemo(() => {
    const verbs = new Set<string>();
    executables.forEach(executable => {
      verbs.add(executable.verb);
      executable.verbAliases?.forEach(alias => verbs.add(alias as string));
    });
    return Array.from(verbs).sort();
  }, [executables]);

  // Handle filter changes
  const handleFilterChange = useCallback((updates: Partial<FilterState>) => {
    setFilterState(prev => {
      const newState = { ...prev, ...updates };

      // Clear dependent filters when parent changes
      if ('workspace' in updates && updates.workspace !== prev.workspace) {
        newState.namespace = '';
      }

      return newState;
    });
  }, []);

  const handleClearAll = useCallback(() => {
    setFilterState({
      search: '',
      tags: [],
      workspace: '',
      namespace: '',
      verb: '',
      visibility: '',
      type: '',
    });
  }, []);

  const activeFilterCount = useMemo(() => {
    let count = 0;
    if (filterState.search) count++;
    if (filterState.tags.length > 0) count++;
    if (filterState.workspace) count++;
    if (filterState.namespace) count++;
    if (filterState.verb) count++;
    if (filterState.visibility) count++
    if (filterState.type) count++;
    return count;
  }, [filterState]);

  return {
    executables: filteredExecutables,
    isLoading,
    error,

    // Filter state and handlers
    filterState,
    onFilterChange: handleFilterChange,
    onClearAll: handleClearAll,
    activeFilterCount,

    // Options for dropdowns
    workspaceOptions,
    namespaceOptions,
    tagOptions,
    verbOptions,

    // Actions
    refresh: refreshExecutables,
  } as const;
}