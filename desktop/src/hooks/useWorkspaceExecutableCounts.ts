import { useQueries } from "@tanstack/react-query";
import { useMemo } from "react";
import { invoke } from "@tauri-apps/api/core";
import { EnrichedWorkspace } from "../types/workspace";
import { EnrichedExecutable } from "../types/executable";

interface WorkspaceExecutableCount {
  workspace: string;
  count: number;
  isLoading: boolean;
  error: Error | null;
}

/**
 * Custom hook to efficiently fetch and cache executable counts for all workspaces.
 * Uses React Query with intelligent caching to avoid redundant API calls.
 *
 * @param workspaces - Array of workspace objects to count executables for
 * @returns Map of workspace name to executable count, plus loading/error states
 */
export function useWorkspaceExecutableCounts(workspaces: EnrichedWorkspace[]) {
  // Create queries for each workspace
  const executableQueries = useQueries({
    queries: workspaces.map((workspace) => ({
      queryKey: ["workspace-executable-count", workspace.name],
      queryFn: async (): Promise<number> => {
        try {
          const executables = await invoke<EnrichedExecutable[]>("list_executables", {
            workspace: workspace.name,
          });
          return executables.length;
        } catch (error) {
          console.warn(`Failed to fetch executables for workspace ${workspace.name}:`, error);
          return 0;
        }
      },
      staleTime: 5 * 60 * 1000, // 5 minutes - data is considered fresh for 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes - keep in cache for 10 minutes after last use
      refetchOnWindowFocus: false, // Don't refetch on window focus for performance
      retry: 1, // Only retry once on failure
    })),
  });

  // Transform query results into a more usable format
  const workspaceExecutableCounts = useMemo(() => {
    const countsMap = new Map<string, WorkspaceExecutableCount>();

    workspaces.forEach((workspace, index) => {
      const query = executableQueries[index];
      countsMap.set(workspace.name, {
        workspace: workspace.name,
        count: query.data ?? 0,
        isLoading: query.isLoading,
        error: query.error,
      });
    });

    return countsMap;
  }, [workspaces, executableQueries]);

  // Overall loading state - true if any workspace is still loading
  const isLoading = executableQueries.some(query => query.isLoading);

  // Check if there are any errors
  const hasErrors = executableQueries.some(query => query.error);

  // Get total executable count across all workspaces
  const totalCount = useMemo(() => {
    return Array.from(workspaceExecutableCounts.values())
      .reduce((sum, { count }) => sum + count, 0);
  }, [workspaceExecutableCounts]);

  /**
   * Get executable count for a specific workspace
   */
  const getCountForWorkspace = (workspaceName: string): number => {
    return workspaceExecutableCounts.get(workspaceName)?.count ?? 0;
  };

  /**
   * Get loading state for a specific workspace
   */
  const isLoadingForWorkspace = (workspaceName: string): boolean => {
    return workspaceExecutableCounts.get(workspaceName)?.isLoading ?? false;
  };

  /**
   * Get error state for a specific workspace
   */
  const getErrorForWorkspace = (workspaceName: string): Error | null => {
    return workspaceExecutableCounts.get(workspaceName)?.error ?? null;
  };

  return {
    // Map of all workspace counts
    workspaceExecutableCounts,

    // Convenience methods
    getCountForWorkspace,
    isLoadingForWorkspace,
    getErrorForWorkspace,

    // Overall states
    isLoading,
    hasErrors,
    totalCount,
  };
}