import { useQuery, useQueryClient } from "@tanstack/react-query";
import React from "react";
import { EnrichedExecutable } from "../types/executable.ts";
import { invoke } from "@tauri-apps/api/core";

export function useExecutable(executableRef: string) {
  const queryClient = useQueryClient();
  const [currentExecutable, setCurrentExecutable] =
    React.useState<EnrichedExecutable | null>(null);

  const {
    data: executable,
    isLoading: isExecutableLoading,
    error: executableError,
  } = useQuery({
    queryKey: ["executable", executableRef],
    queryFn: async () => {
      if (!executableRef) return null;
      return await invoke<EnrichedExecutable>("get_executable", {
        executableRef: executableRef,
      });
    },
    enabled: !!executableRef,
  });

  // Update current executable when we have new data
  React.useEffect(() => {
    if (executable) {
      setCurrentExecutable(executable);
    }
  }, [executable]);

  const refreshExecutable = () => {
    if (executableRef) {
      queryClient.invalidateQueries({
        queryKey: ["executable", executableRef],
      });
    }
  };

  return {
    executable: currentExecutable,
    isExecutableLoading,
    executableError,
    refreshExecutable,
  };
}

export function useExecutables(
  workspace: string | null,
  namespace: string | null,
  tags: string[] | null,
  verb: string | null,
  filter: string | null,
) {
  const queryClient = useQueryClient();

  const normalizedParams = React.useMemo(() => {
    return {
      workspace: workspace || undefined,
      namespace: namespace || undefined,
      tags: tags && tags.length > 0 ? tags : undefined,
      verb: verb || undefined,
      filter: filter || undefined,
    };
  }, [workspace, namespace, tags, verb, filter]);

  const queryKey = React.useMemo(() => {
    return ["executables", normalizedParams];
  }, [normalizedParams]);

  const {
    data: executables,
    isLoading: isExecutablesLoading,
    error: executablesError,
  } = useQuery({
    queryKey,
    queryFn: async () => {
      console.log('Fetching executables with params:', normalizedParams);
      return await invoke<EnrichedExecutable[]>("list_executables", {
        workspace: normalizedParams.workspace,
        namespace: normalizedParams.namespace,
        tags: normalizedParams.tags,
        verb: normalizedParams.verb,
        filter: normalizedParams.filter,
      });
    },
  });

  const refreshExecutables = () => {
    void queryClient.invalidateQueries({
      queryKey,
    });
  };

  return {
    executables: executables || [],
    isExecutablesLoading,
    executablesError,
    refreshExecutables,
  };
}