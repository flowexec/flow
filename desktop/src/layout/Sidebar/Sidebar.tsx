import { Group, Image, NavLink, Stack } from "@mantine/core";
import {
  IconDatabase,
  IconFolders,
  IconLogs,
  IconSettings,
  IconPlayerPlay,
} from "@tabler/icons-react";
import { useCallback } from "react";
import { Link, useLocation } from "wouter";
import styles from "./Sidebar.module.css";
import iconImage from "/logo-dark.png";

export function Sidebar() {
  const [location, setLocation] = useLocation();

  const navigateToWorkspaces = useCallback(() => {
    setLocation(`/workspaces`);
  }, [setLocation]);

  const navigateToExecutables = useCallback(() => {
    setLocation(`/executables`);
  }, [setLocation]);

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
        <Group gap="xs" mt="md">
          <NavLink
            label="Workspaces"
            leftSection={<IconFolders size={16} />}
            active={location.startsWith("/workspaces")}
            variant="filled"
            onClick={navigateToWorkspaces}
          />

          <NavLink
            label="Executables"
            leftSection={<IconPlayerPlay size={16} />}
            active={location.startsWith("/executables")}
            variant="filled"
            onClick={navigateToExecutables}
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

      </Stack>
    </div>
  );
}
