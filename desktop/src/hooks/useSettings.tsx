import { useLocalStorage } from "@mantine/hooks";
import { createContext, ReactNode, useContext } from "react";
import { ThemeName } from "../theme/types";

interface Settings {
  workspaceApp: string;
  executableApp: string;
  theme: ThemeName;
}

interface SettingsContextType {
  settings: Settings;
  updateWorkspaceApp: (value: string) => void;
  updateExecutableApp: (value: string) => void;
  updateTheme: (value: ThemeName) => void;
}

const defaultSettings: Settings = {
  workspaceApp: "",
  executableApp: "",
  theme: "everforest",
};

const SettingsContext = createContext<SettingsContextType | undefined>(
  undefined,
);

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [workspaceApp, setWorkspaceApp] = useLocalStorage<string>({
    key: "workspaceApp",
    defaultValue: defaultSettings.workspaceApp,
  });

  const [executableApp, setExecutableApp] = useLocalStorage<string>({
    key: "executableApp",
    defaultValue: defaultSettings.executableApp,
  });

  const [theme, setTheme] = useLocalStorage<ThemeName>({
    key: "theme",
    defaultValue: defaultSettings.theme,
  });

  const value = {
    settings: {
      workspaceApp,
      executableApp,
      theme,
    },
    updateWorkspaceApp: setWorkspaceApp,
    updateExecutableApp: setExecutableApp,
    updateTheme: setTheme,
  };

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  const context = useContext(SettingsContext);
  if (context === undefined) {
    throw new Error("useSettings must be used within a SettingsProvider");
  }
  return context;
}
