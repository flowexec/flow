import { useQueries, useQuery } from "@tanstack/react-query";
import { invoke } from "@tauri-apps/api/core";
import { useMemo } from "react";
import { EnrichedExecutable } from "../types/executable";
import { EnrichedWorkspace } from "../types/workspace";

interface WorkspaceExecutableCount {
  workspace: string;
  count: number;
  isLoading: boolean;
  error: Error | null;
}

export function useWorkspaceExecutableCounts(workspaces: EnrichedWorkspace[]) {
  const executableQueries = useQueries({
    queries: workspaces.map((workspace) => ({
      queryKey: ["workspace-executable-count", workspace.name],
      queryFn: async (): Promise<number> => {
        try {
          const executables = await invoke<EnrichedExecutable[]>(
            "list_executables",
            {
              workspace: workspace.name,
            }
          );
          return executables.length;
        } catch (error) {
          console.warn(
            `Failed to fetch executables for workspace ${workspace.name}:`,
            error
          );
          return 0;
        }
      },
      retry: 1,
    })),
  });

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

  const isLoading = executableQueries.some((query) => query.isLoading);
  const hasErrors = executableQueries.some((query) => query.error);
  const totalCount = useMemo(() => {
    return Array.from(workspaceExecutableCounts.values()).reduce(
      (sum, { count }) => sum + count,
      0
    );
  }, [workspaceExecutableCounts]);

  const getCountForWorkspace = (workspaceName: string): number => {
    return workspaceExecutableCounts.get(workspaceName)?.count ?? 0;
  };

  const isLoadingForWorkspace = (workspaceName: string): boolean => {
    return workspaceExecutableCounts.get(workspaceName)?.isLoading ?? false;
  };

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

export function useWorkspaceExecutableCount(
  workspace: EnrichedWorkspace | string | null | undefined
) {
  const workspaceName = typeof workspace === "string" ? workspace : workspace?.name;

  const query = useQuery<number, Error>({
    queryKey: ["workspace-executable-count", workspaceName],
    queryFn: async (): Promise<number> => {
      if (!workspaceName) return 0;
      try {
        const executables = await invoke<EnrichedExecutable[]>("list_executables", {
          workspace: workspaceName,
        });
        return executables.length;
      } catch (error) {
        console.warn(`Failed to fetch executables for workspace ${workspaceName}:`, error);
        return 0;
      }
    },
    enabled: !!workspaceName,
    retry: 1,
  });

  return {
    count: query.data ?? 0,
    isLoading: !!workspaceName && query.isLoading,
    error: query.error ?? null,
  };
}
