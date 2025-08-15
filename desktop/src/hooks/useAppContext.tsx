import React from "react";
import { createContext, useContext, useState, useEffect } from "react";
import { Config } from "../types/generated/config";
import { EnrichedWorkspace } from "../types/workspace";
import { EnrichedExecutable } from "../types/executable";
import { useConfig } from "./useConfig";
import { useWorkspaces } from "./useWorkspace";
import { useExecutables } from "./useExecutable";
import { invoke } from "@tauri-apps/api/core";
import { useQuery } from "@tanstack/react-query";

interface AppContextType {
    config: Config | null;
    selectedWorkspace: string | null;
    setSelectedWorkspace: (workspaceName: string | null) => void;
    workspaces: EnrichedWorkspace[];
    executables: EnrichedExecutable[];
    isLoading: boolean;
    hasError: Error | null;
    refreshAll: () => void;
}

export const AppContext = createContext<AppContextType | undefined>(undefined);

export function useFlowBinaryCheck() {
    const { data: isFlowBinaryAvailable, isLoading: isCheckingBinary, error: binaryCheckError } = useQuery({
        queryKey: ["flowBinaryCheck"],
        queryFn: async () => {
            try {
                await invoke("check_flow_binary");
                return true;
            } catch (error) {
                console.error(error);
                throw new Error("flow CLI not found or not executable");
            }
        },
        retry: false,
        refetchOnWindowFocus: false,
    });

    return {
        isFlowBinaryAvailable,
        isCheckingBinary,
        binaryCheckError,
    };
}

export function AppProvider({ children }: { children: React.ReactNode }) {
    const { isCheckingBinary, binaryCheckError } = useFlowBinaryCheck();
    const enabled = !binaryCheckError && !isCheckingBinary;

    const { config, isConfigLoading, configError, refreshConfig } = useConfig(enabled);

    const { workspaces, isWorkspacesLoading, workspacesError, refreshWorkspaces } = useWorkspaces(enabled);
    const [selectedWorkspace, setSelectedWorkspace] = useState<string | null>(null);

    useEffect(() => {
        if (config?.currentWorkspace && workspaces && workspaces.length > 0) {
            // Only set if we don't have a selected workspace or if the config workspace exists
            const configWorkspaceExists = workspaces.some(w => w.name === config.currentWorkspace);
            if (!selectedWorkspace && configWorkspaceExists) {
                setSelectedWorkspace(config.currentWorkspace);
            }
        }
    }, [config, workspaces, selectedWorkspace]);

    const { executables, isExecutablesLoading, executablesError, refreshExecutables } = useExecutables(selectedWorkspace, enabled);

    const isLoading =
        isConfigLoading || isWorkspacesLoading || isExecutablesLoading;
    const hasError = configError || workspacesError || executablesError;
    if (hasError) {
        console.error("Error", hasError);
    }

    const refreshAll = () => {
        refreshConfig();
        refreshWorkspaces();
        refreshExecutables();
    };

    // If flow binary is not available, return early with error state
    if (binaryCheckError) {
        return (
            <AppContext.Provider value={{
                config: null,
                workspaces: [],
                selectedWorkspace,
                setSelectedWorkspace,
                executables: [],
                isLoading: false,
                hasError: binaryCheckError,
                refreshAll: () => {},
            }}>
                {children}
            </AppContext.Provider>
        );
    }

    // If still checking binary, show loading state
    if (isCheckingBinary) {
        return (
            <AppContext.Provider value={{
                config: null,
                workspaces: [],
                selectedWorkspace,
                setSelectedWorkspace,
                executables: [],
                isLoading: true,
                hasError: null,
                refreshAll: () => {},
            }}>
                {children}
            </AppContext.Provider>
        );
    }

    return (
        <AppContext.Provider value={{
            config: config || null,
            workspaces: workspaces || [],
            selectedWorkspace,
            setSelectedWorkspace,
            executables,
            isLoading,
            hasError,
            refreshAll,
        }}>
            {children}
        </AppContext.Provider>
    );
}

export function useAppContext() {
    const context = useContext(AppContext);
    if (context === undefined) {
        throw new Error('useAppContext must be used within an AppProvider');
    }
    return context;
}