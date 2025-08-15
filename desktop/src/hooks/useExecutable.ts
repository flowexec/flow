import {useQuery, useQueryClient} from "@tanstack/react-query";
import React from "react";
import {EnrichedExecutable} from "../types/executable.ts";
import {invoke} from "@tauri-apps/api/core";

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

export function useExecutables(selectedWorkspace: string | null, enabled: boolean = true) {
    const queryClient = useQueryClient();

    const {
        data: executables,
        isLoading: isExecutablesLoading,
        error: executablesError,
    } = useQuery({
        queryKey: ["executables", selectedWorkspace],
        queryFn: async () => {
            if (!selectedWorkspace) return [];
            return await invoke<EnrichedExecutable[]>("list_executables", {
                workspace: selectedWorkspace,
            });
        },
        enabled: enabled && !!selectedWorkspace, // Only run when workspace is selected AND enabled
    });

    const refreshExecutables = () => {
        if (selectedWorkspace) {
            queryClient.invalidateQueries({
                queryKey: ["executables", selectedWorkspace],
            });
        }
    };

    return {
        executables: executables || [],
        isExecutablesLoading,
        executablesError,
        refreshExecutables,
    };
}
