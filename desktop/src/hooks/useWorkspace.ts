import {useQuery, useQueryClient} from "@tanstack/react-query";
import {invoke} from "@tauri-apps/api/core";
import {EnrichedWorkspace} from "../types/workspace.ts";

export function useWorkspace(workspaceName: string, enabled: boolean = true) {
    const queryClient = useQueryClient();

    const {
        data: workspace,
        isLoading: isWorkspaceLoading,
        error: workspaceError,
    } = useQuery({
        queryKey: ["workspace", workspaceName],
        queryFn: async () => {
            return await invoke<EnrichedWorkspace>("get_workspace", {
                name: workspaceName,
            });
        },
        enabled,
    });

    const refreshWorkspace = () => {
        if (workspaceName) {
            queryClient.invalidateQueries({queryKey: ["workspace", workspaceName]});
        }
    };

    return {
        workspace,
        workspaceError,
        isWorkspaceLoading,
        refreshWorkspace,
    };
}

export function useWorkspaces(enabled: boolean = true) {
    const queryClient = useQueryClient();

    const {
        data: workspaces,
        isLoading: isWorkspacesLoading,
        error: workspacesError,
    } = useQuery({
        queryKey: ["workspaces"],
        queryFn: async () => {
            return await invoke<EnrichedWorkspace[]>("list_workspaces");
        },
        enabled,
    });

    const refreshWorkspaces = () => {
        queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    };

    return {
        workspaces,
        isWorkspacesLoading,
        workspacesError,
        refreshWorkspaces,
    };
}
