import {useQuery, useQueryClient} from "@tanstack/react-query";
import {invoke} from "@tauri-apps/api/core";
import {Config} from "../types/generated/config.ts";

export function useConfig(enabled: boolean = true) {
    const queryClient = useQueryClient();

    const {
        data: config,
        isLoading: isConfigLoading,
        error: configError,
    } = useQuery({
        queryKey: ["config"],
        queryFn: async () => {
            return await invoke<Config>("get_config");
        },
        enabled,
    });

    const refreshConfig = () => {
        queryClient.invalidateQueries({ queryKey: ["config"] });
    };

    const updateTheme = async (theme: string) => {
        try {
            await invoke("set_config_theme", { theme });
            refreshConfig();
        } catch (error) {
            throw new Error(`Failed to update theme: ${error}`);
        }
    };

    const updateWorkspaceMode = async (mode: string) => {
        try {
            await invoke("set_config_workspace_mode", { mode });
            refreshConfig();
        } catch (error) {
            throw new Error(`Failed to update workspace mode: ${error}`);
        }
    };

    const updateLogMode = async (mode: string) => {
        try {
            await invoke("set_config_log_mode", { mode });
            refreshConfig();
        } catch (error) {
            throw new Error(`Failed to update log mode: ${error}`);
        }
    };

    const updateNamespace = async (namespace: string) => {
        try {
            await invoke("set_config_namespace", { namespace });
            refreshConfig();
        } catch (error) {
            throw new Error(`Failed to update namespace: ${error}`);
        }
    };

    const updateCurrentWorkspace = async (workspace: string) => {
        try {
            await invoke("set_workspace", { name: workspace, fixed: false });
            refreshConfig();
        } catch (error) {
            throw new Error(`Failed to update current workspace: ${error}`);
        }
    };

    const updateDefaultTimeout = async (timeout: string) => {
        try {
            await invoke("set_config_timeout", { timeout });
            refreshConfig();
        } catch (error) {
            throw new Error(`Failed to update default timeout: ${error}`);
        }
    };

    return {
        config,
        isConfigLoading,
        configError,
        refreshConfig,
        updateTheme,
        updateWorkspaceMode,
        updateLogMode,
        updateNamespace,
        updateCurrentWorkspace,
        updateDefaultTimeout,
    };
}
