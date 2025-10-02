import {
  ActionIcon,
  Box,
  Button, ButtonGroup,
  Flex,
  Group,
  Image,
  Menu,
  NavLink,
  Stack,
  Text,
  Tooltip,
} from "@mantine/core";
import {
  IconDatabase,
  IconFolders,
  IconLogs,
  IconSettings,
  IconTerminal2,
  IconInfoCircle,
  IconReload,
  IconLayoutSidebarLeftExpand,
  IconLayoutSidebarLeftCollapse,
  IconStarFilled,
  IconStar,
  IconTools,
  IconArrowRight,
  IconChevronCompactRight,
  IconChevronRight,
  IconShield,
  IconShieldLockFilled,
  IconTemplate,
  IconShieldLock,
} from "@tabler/icons-react";
import { useCallback } from "react";
import { Link, useLocation } from "wouter";
import { useAppContext } from "../../hooks/useAppContext.tsx";
import styles from "./Sidebar.module.css";
import iconImage from "/icon.png";

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const [location, setLocation] = useLocation();
  const { config, selectedWorkspace } = useAppContext();
  const currentWorkspaceName = selectedWorkspace || config?.currentWorkspace || "—";
  const currentNamespace = config?.currentNamespace || "—";

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
    <Stack justify="space-between" className={`${styles.sidebar} ${collapsed ? styles.sidebarCollapsed : ""}`}>
      <Box>
        <Flex justify="space-between" align={ collapsed ? "center" : "flex-end" } direction={ collapsed ? "column-reverse" : "column" }>
          <ActionIcon
            variant="transparent"
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
            onClick={onToggle}
            color="bodyLight"
          >
            {collapsed ? <IconLayoutSidebarLeftExpand size={16} /> : <IconLayoutSidebarLeftCollapse size={16} />}
          </ActionIcon>
          <Link to="/" className={styles.sidebar__logo}>
            <Image src={iconImage} alt="flow" fit="contain" />
          </Link>
        </Flex>

        <Stack gap="xs" mt="xl" justify="stretch" align={ collapsed ? "center" : "flex-start" }>
          <NavLink
            label="Favorites"
            leftSection={<IconStar size={16} />}
            active={location.startsWith("/favorites")}
            variant="filled"
            onClick={navigateToWorkspaces}
          />
          <NavLink
            label="Workspaces"
            leftSection={<IconFolders size={16} />}
            active={location.startsWith("/workspaces")}
            variant="filled"
            onClick={navigateToWorkspaces}
          />

          <NavLink
            label="Executables"
            leftSection={<IconTerminal2 size={16} />}
            active={location.startsWith("/executables")}
            variant="filled"
            onClick={navigateToExecutables}
          />
        </Stack>
      </Box>
      <Box>
        {collapsed ? (
          <Group justify="center" gap={8}>
            <Tooltip label={`Workspace: ${currentWorkspaceName} | Namespace: ${currentNamespace}`} position="right" openDelay={300}>
              <ActionIcon variant="transparent" size="sm" aria-label="Context">
                <IconInfoCircle size={14} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Sync" position="right" openDelay={300}>
              <ActionIcon variant="light" size="sm" aria-label="Sync workspaces and executables">
                <IconReload size={14} />
              </ActionIcon>
            </Tooltip>
          </Group>
        ) : (

          <Stack gap="xs" align="stretch" justify="center">

            <Box className={styles.sidebar__context}>
              <Group gap="xs" justify="space-between" mb="xs">
                <Text size="xs" fw={700} c="dimmed">Context</Text>
                <IconInfoCircle size={16} />
              </Group>
              <Group>
                <Text size="xs" c="dimmed">Workspace</Text>
                <Text size="xs" truncate>{currentWorkspaceName}</Text>
              </Group>
              <Group>
                <Text size="xs" c="dimmed" mt={6}>Namespace</Text>
                <Text size="xs" truncate>{currentNamespace}</Text>
              </Group>
            </Box>
            <ButtonGroup style={{ alignSelf: 'center' }}>
              <Button
                leftSection={<IconReload size={12} />}
                size="compact-xs" variant="transparent" justify="start">Sync</Button>
              <Menu shadow="md" position="top-start" offset={15} width={200} withArrow>
                <Menu.Target>
                  <Button
                    leftSection={<IconTools size={12} />}
                    size="compact-xs"
                    variant="transparent"
                    onClick={navigateToSettings}
                    justify="start"
                  >Tools</Button>
                </Menu.Target>

                <Menu.Dropdown>
                  <Menu.Item leftSection={<IconShieldLock size={14} />}>
                    Vault
                  </Menu.Item>
                  <Menu.Item leftSection={<IconDatabase size={14} />}>
                    Cache Store
                  </Menu.Item>
                  <Menu.Item leftSection={<IconTemplate size={14} />}>
                    Templates
                  </Menu.Item>
                  <Menu.Item
                    leftSection={<IconLogs size={14} />}
                  >
                    Logs
                  </Menu.Item>
                </Menu.Dropdown>
              </Menu>
              <Button
                leftSection={<IconSettings size={12} />}
                size="compact-xs"
                variant="transparent"
                onClick={navigateToSettings}
                justify="start"
              >Settings</Button>
            </ButtonGroup>
          </Stack>
        )}
      </Box>
    </Stack>
  );
}
