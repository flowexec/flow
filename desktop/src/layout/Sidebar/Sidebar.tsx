import { Group, Image, NavLink, Stack } from "@mantine/core";
import { ExecutableTree } from "./ExecutableTree/ExecutableTree";
import styles from "./Sidebar.module.css";
import { WorkspaceSelector } from "./WorkspaceSelector/WorkspaceSelector";
import iconImage from "/logo-dark.png";
import {
  IconDatabase,
  IconFolders,
  IconLogs,
  IconSettings,
} from "@tabler/icons-react";
import { Link, useLocation } from "wouter";
import { useAppContext } from "../../hooks/useAppContext.tsx";
import { useCallback } from "react";

export function Sidebar() {
  const [location, setLocation] = useLocation();
  const { executables, selectedWorkspace } = useAppContext();

  const navigateToWorkspace = useCallback(() => {
    setLocation(`/workspace/${selectedWorkspace || ""}`);
  }, [setLocation, selectedWorkspace]);

  const navigateToLogs = useCallback(() => {
    setLocation("/logs");
  }, [setLocation]);

  const navigateToCache = useCallback(() => {
    setLocation("/cache");
  }, [setLocation]);

  const navigateToVault = useCallback(() => {
    setLocation("/vault");
  }, [setLocation]);

  const navigateToSettings = useCallback(() => {
    setLocation("/settings");
  }, [setLocation]);

  return (
    <div className={styles.sidebar}>
      <Link to="/" className={styles.sidebar__logo}>
        <Image src={iconImage} alt="flow" fit="contain" />
      </Link>
      <Stack gap="xs">
        <WorkspaceSelector />

        <Group gap="xs" mt="md">
          <NavLink
            label="Workspace"
            leftSection={<IconFolders size={16} />}
            active={location.startsWith("/workspace")}
            variant="filled"
            onClick={navigateToWorkspace}
          />

          <NavLink
            label="Logs"
            leftSection={<IconLogs size={16} />}
            active={location.startsWith("/logs")}
            variant="filled"
            onClick={navigateToLogs}
          />

          <NavLink
            label="Data"
            leftSection={<IconDatabase size={16} />}
            variant="filled"
            childrenOffset={28}
          >
            <NavLink
              label="Cache"
              variant="filled"
              active={location.startsWith("/cache")}
              onClick={navigateToCache}
            />
            <NavLink
              label="Vault"
              variant="filled"
              active={location.startsWith("/vault")}
              onClick={navigateToVault}
            />
          </NavLink>

          <NavLink
            label="Settings"
            leftSection={<IconSettings size={16} />}
            active={location.startsWith("/settings")}
            variant="filled"
            onClick={navigateToSettings}
          />
        </Group>

        {executables && executables.length > 0 && <ExecutableTree />}
      </Stack>
    </div>
  );
}
