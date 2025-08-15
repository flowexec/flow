import { 
  Select, 
  Stack,
  TextInput, 
  Title,
  LoadingOverlay,
  Alert,
  Paper,
} from "@mantine/core";
import { IconInfoCircle } from "@tabler/icons-react";
import { useConfig } from "../../hooks/useConfig";
import { useSettings } from "../../hooks/useSettings";
import { useNotifier } from "../../hooks/useNotifier";
import { ThemeName } from "../../theme/types";
import { NotificationType } from "../../types/notification";
import { SettingRow, SettingSection } from "../../components/Settings";
import styles from "./Settings.module.css";
import { useState, useEffect } from "react";
import { PageWrapper } from "../../components/PageWrapper";

const themeOptions = [
  { value: "everforest", label: "Default" },
  { value: "dark", label: "Dark" },
  { value: "dracula", label: "Dracula" },
  { value: "light", label: "Light" },
  { value: "tokyo-night", label: "Tokyo Night" },
];

const workspaceModeOptions = [
  { value: "fixed", label: "Fixed" },
  { value: "dynamic", label: "Dynamic" },
];

const logModeOptions = [
  { value: "hidden", label: "Hidden" },
  { value: "json", label: "JSON" },
  { value: "logfmt", label: "Log Format" },
  { value: "text", label: "Text" },
];

export function Settings() {
  const { settings, updateWorkspaceApp, updateExecutableApp, updateTheme } =
      useSettings();
  const {
    config,
    isConfigLoading,
    refreshConfig,
    configError,
    updateTheme: updateConfigTheme,
    updateWorkspaceMode,
    updateLogMode,
    updateNamespace,
    updateCurrentWorkspace,
    updateDefaultTimeout,
  } = useConfig();
  const { setNotification } = useNotifier();
  const [namespaceInput, setNamespaceInput] = useState<string>('');

  useEffect(() => {
    if (config) {
      setNamespaceInput(config.currentNamespace || '');
    }
  }, [config]);

  if (configError) {
    return (
      <div className={styles.settings}>
        <Title order={2} mb="md">Settings</Title>
        <Alert variant="light" color="red" icon={<IconInfoCircle />}>
          Error loading configuration: {configError.message}
        </Alert>
      </div>
    );
  }

  async function handleThemeChange(value: string) {
    updateTheme(value as ThemeName);
    return updateConfigTheme(value).then(() => {
      refreshConfig();
      setNotification({
        type: NotificationType.Success,
        title: 'Theme updated',
        message: 'Theme has been successfully updated',
        autoClose: true,
      });
    }).catch((error) => {
      setNotification({
        type: NotificationType.Error,
        title: 'Error updating theme',
        message: error.message,
        autoClose: true,
      });
    });
  }

  async function handleLogModeChange(value: string) {
    return updateLogMode(value).then(() => {
      refreshConfig();
      setNotification({
        type: NotificationType.Success,
        title: 'Log mode updated',
        message: 'Log mode has been successfully updated',
        autoClose: true,
      });
    }).catch((error) => {
      setNotification({
        type: NotificationType.Error,
        title: 'Error updating log mode',
        message: error.message,
        autoClose: true,
      });
    });
  }

  async function handleDefaultTimeoutChange(value: string){
    return updateDefaultTimeout(value).then(() => {
      refreshConfig();
      setNotification({
        type: NotificationType.Success,
        title: 'Default timeout updated',
        message: 'Default timeout has been successfully updated',
        autoClose: true,
      });
    }).catch((error) => {
      setNotification({
        type: NotificationType.Error,
        title: 'Error updating default timeout',
        message: error.message,
        autoClose: true,
      });
    });
  }

  async function handleCurrentWorkspaceChange(value: string){
    return updateCurrentWorkspace(value).then(() => {
      refreshConfig();
      setNotification({
        type: NotificationType.Success,
        title: 'Current workspace updated',
        message: 'Current workspace has been successfully updated',
        autoClose: true,
      });
    }).catch((error) => {
      setNotification({
        type: NotificationType.Error,
        title: 'Error updating current workspace',
        message: error.message,
        autoClose: true,
      });
    });
  }

  async function handleWorkspaceModeChange(value: string){
    return updateWorkspaceMode(value).then(() => {
      refreshConfig();
      setNotification({
        type: NotificationType.Success,
        title: 'Workspace mode updated',
        message: 'Workspace mode has been successfully updated',
        autoClose: true,
      });
    }).catch((error) => {
      setNotification({
        type: NotificationType.Error,
        title: 'Error updating workspace mode',
        message: error.message,
        autoClose: true,
      });
    });
  }

  async function handleNamespaceChange(value: string){
    return updateNamespace(value).then(() => {
      refreshConfig();
      setNotification({
        type: NotificationType.Success,
        title: 'Namespace updated',
        message: 'Namespace has been successfully updated',
        autoClose: true,
      });
    }).catch((error) => {
      setNotification({
        type: NotificationType.Error,
        title: 'Error updating namespace',
        message: error.message,
        autoClose: true,
      });
    });
  }

  function handleNamespaceSubmit() {
    if (namespaceInput !== (config?.currentNamespace || '')) {
      handleNamespaceChange(namespaceInput);
    }
  }

  return (
    <PageWrapper>
      <div className={styles.settings}>
        <LoadingOverlay visible={isConfigLoading} />

        <Title order={2} mb="xl">
          Settings
        </Title>

        <Stack gap={0}>
          <Paper className={styles.settingCard} mb="lg">
            <SettingRow
              label="Theme"
              description="Choose your preferred color theme"
            >
              <Select
                size="sm"
                value={config?.theme || "everforest"}
                onChange={(value) => value && handleThemeChange(value)}
                data={themeOptions}
                variant="filled"
              />
            </SettingRow>

            <SettingRow
              label="Default Log Mode"
              description="Default logging format for executable runs"
            >
              <Select
                size="sm"
                value={config?.defaultLogMode || "text"}
                onChange={(value) => value && handleLogModeChange(value)}
                data={logModeOptions}
                variant="filled"
              />
            </SettingRow>

            <SettingRow
              label="Default Timeout"
              description="Default timeout for executable runs"
            >
              <TextInput
                size="sm"
                value={config?.defaultTimeout || ""}
                onChange={(e) => handleDefaultTimeoutChange(e.currentTarget.value)}
                placeholder="e.g., 30s, 5m, 1h"
                variant="filled"
              />
            </SettingRow>
          </Paper>

          <SettingSection title="Workspace">
            <SettingRow
              label="Current Workspace"
              description="The currently active workspace"
            >
              <Select
                size="sm"
                value={config?.currentWorkspace || ""}
                onChange={(value) => value && handleCurrentWorkspaceChange(value)}
                data={Object.keys(config?.workspaces || {}).map(name => ({ value: name, label: name }))}
                placeholder="Select workspace"
                variant="filled"
              />
            </SettingRow>

            <SettingRow
              label="Workspace Mode"
              description="Dynamic mode changes global workspace when switching in sidebar"
            >
              <Select
                size="sm"
                value={config?.workspaceMode || "dynamic"}
                onChange={(value) => value && handleWorkspaceModeChange(value)}
                data={workspaceModeOptions}
                variant="filled"
              />
            </SettingRow>

            <SettingRow
              label="Current Namespace"
              description="Active namespace for executable discovery"
            >
              <TextInput
                size="sm"
                value={namespaceInput}
                onChange={(e) => setNamespaceInput(e.currentTarget.value)}
                onBlur={handleNamespaceSubmit}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    handleNamespaceSubmit();
                  }
                }}
                placeholder="Enter namespace"
                variant="filled"
                spellCheck={false}
              />
            </SettingRow>
          </SettingSection>

          <SettingSection title="External Applications">
            <SettingRow
              label="Workspace Command"
              description="Command to open workspace directories"
            >
              <TextInput
                size="sm"
                value={settings.workspaceApp}
                onChange={(event) => updateWorkspaceApp(event.currentTarget.value)}
                placeholder="System default"
                variant="filled"
                spellCheck={false}
              />
            </SettingRow>

            <SettingRow
              label="Executable Command"
              description="Command to open flow files"
            >
              <TextInput
                size="sm"
                value={settings.executableApp}
                onChange={(event) => updateExecutableApp(event.currentTarget.value)}
                placeholder="System default"
                variant="filled"
                spellCheck={false}
              />
            </SettingRow>
          </SettingSection>
        </Stack>
      </div>
    </PageWrapper>
  );
}
