import { ActionIcon, Group, Menu, Badge, Table, Text, Stack } from "@mantine/core";
import {
  IconDotsVertical,
  IconExternalLink,
  IconFile,
  IconPlayerPlayFilled,
  IconOctagon,
  IconSettingsAutomation,
  IconProgressX,
  IconProgressDown,
  IconRefresh,
  IconCircleCheckFilled,
  IconWindowMaximize,
  IconCirclePlus,
  IconReload,
  IconBlocks,
  IconStar,
} from "@tabler/icons-react";
import clsx from "clsx";
import type { EnrichedExecutable } from "../../types/executable";
import { GetUIVerbType, DeactivationVerbType, ConfigurationVerbType, DestructionVerbType, RetrievalVerbType, UpdateVerbType, ValidationVerbType, LaunchVerbType, CreationVerbType, RestartVerbType, BuildVerbType } from "../../types/executable";
import { stringToColor } from "../../utils/colors";
import { openPath } from "@tauri-apps/plugin-opener";
import { useNotifier } from "../../hooks/useNotifier";
import { NotificationType } from "../../types/notification";
import { shortenCleanDescription } from "../../utils/text";
import classes from "./Executables.module.css";

export enum ViewMode {
  List = "list",
  Card = "card",
}

function IconForVerbType(verbType: string | null) {
  switch (verbType) {
    case DeactivationVerbType:
      return <IconOctagon size={18} className={classes.verbIcon} />;
    case ConfigurationVerbType:
      return <IconSettingsAutomation size={18} className={classes.verbIcon} />;
    case DestructionVerbType:
      return <IconProgressX size={18} className={classes.verbIcon} />;
    case RetrievalVerbType:
      return <IconProgressDown size={18} className={classes.verbIcon} />;
    case UpdateVerbType:
      return <IconRefresh size={18} className={classes.verbIcon} />;
    case ValidationVerbType:
      return <IconCircleCheckFilled size={18} className={classes.verbIcon} />;
    case LaunchVerbType:
      return <IconWindowMaximize size={18} className={classes.verbIcon} />;
    case CreationVerbType:
      return <IconCirclePlus size={18} className={classes.verbIcon} />;
    case RestartVerbType:
      return <IconReload size={18} className={classes.verbIcon} />;
    case BuildVerbType:
      return <IconBlocks size={18} className={classes.verbIcon} />;
    default:
      return <IconPlayerPlayFilled size={18} className={classes.verbIcon} />;
  }
}

export function ExecutablesTable({
  items,
  loading,
  hasActiveFilters,
  onRowClick,
  viewMode,
 }: {
  items: EnrichedExecutable[];
  loading: boolean;
  hasActiveFilters: boolean;
  onRowClick: (ref: string) => void;
  viewMode: ViewMode;
}) {
  const { setNotification } = useNotifier();

  const rows = items.map((exec) => (
    <Table.Tr
      key={exec.ref}
      onClick={() => onRowClick(exec.ref)}
      style={{ cursor: 'pointer' }}
      className={classes.tableRow}
    >
      <Table.Td className={clsx(classes.cell, classes.iconCell)}>
        {IconForVerbType(GetUIVerbType(exec))}
      </Table.Td>
      <Table.Td className={classes.cell}>
        <Stack gap="sm">
          <Group justify="space-between" align="flex-start">
            <Text className={classes.executableRef}>
              {exec.ref}
            </Text>

            <Menu shadow="md" width={200} position="bottom-end">
              <Menu.Target>
                <ActionIcon
                  variant="subtle"
                  size="sm"
                  aria-label="Executable actions"
                  onClick={(e) => e.stopPropagation()}
                >
                  <IconDotsVertical size={16} />
                </ActionIcon>
              </Menu.Target>
              <Menu.Dropdown onClick={(e) => e.stopPropagation()}>
                <Menu.Item leftSection={<IconExternalLink size={14} />} onClick={() => onRowClick(exec.ref)}>
                  Open details
                </Menu.Item>
                <Menu.Item leftSection={<IconFile size={14} />} onClick={async () => {
                  try {
                    await openPath(exec.flowfile);
                  } catch (error) {
                    setNotification({ type: NotificationType.Error, title: "Failed to open flowfile", message: String(error) });
                  }
                }}>
                  Open flowfile
                </Menu.Item>
                <Menu.Divider />
                <Menu.Item leftSection={<IconStar size={14} />} onClick={() => {
                  setNotification({ type: NotificationType.Success, title: "Favorited", message: `${exec.ref} saved as favorite` });
                }}>
                  Save as favorite
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </Group>

          {viewMode === ViewMode.Card && exec.fullDescription && (
            <Text size="sm" c="dimmed" style={{ lineHeight: 1.4 }}>
              {shortenCleanDescription(exec.fullDescription, 180)}
            </Text>
          )}

          {viewMode === ViewMode.Card && (
            <Group gap={6} h="lg">
              <Badge
                size="xs"
                variant="light"
                color={getExecutableTypeColor(exec)}
                className={classes.tag}
              >
                {getExecutableTypeLabel(exec)}
              </Badge>
              {exec.tags && exec.tags.length > 0 && exec.tags.map((tag, i) => (
                <Badge
                  key={i}
                  size="xs"
                  variant="outline"
                  color={stringToColor(tag)}
                  className={classes.tag}
                >
                  {tag}
                </Badge>
              ))}
            </Group>
          )}
        </Stack>
      </Table.Td>
    </Table.Tr>
  ));

  return (
    <Table className={classes.table} highlightOnHover>
      <Table.Tbody>
        {rows.length > 0 ? (
          rows
        ) : (
          <Table.Tr>
            <Table.Td>
              <Group justify="center" gap="sm" className={classes.emptyState}>
                <Text c="dimmed" size="sm">
                  {loading
                    ? "Loading executables..."
                    : hasActiveFilters
                      ? "No executables match your filters"
                      : "No executables found"}
                </Text>
              </Group>
            </Table.Td>
          </Table.Tr>
        )}
      </Table.Tbody>
    </Table>
  );
}

function getExecutableTypeColor(exec: EnrichedExecutable): string {
  if (exec.exec) return 'blue.5';
  if (exec.serial) return 'green.5';
  if (exec.parallel) return 'orange.5';
  if (exec.launch) return 'purple.5';
  if (exec.request) return 'teal.5';
  if (exec.render) return 'pink.5';
  return 'gray.5';
}

function getExecutableTypeLabel(exec: EnrichedExecutable): string {
  if (exec.exec) return 'Command';
  if (exec.serial) return 'Serial Workflow';
  if (exec.parallel) return 'Parallel Workflow';
  if (exec.launch) return 'Launch';
  if (exec.request) return 'HTTP Request';
  if (exec.render) return 'Template';
  return 'Unknown';
}