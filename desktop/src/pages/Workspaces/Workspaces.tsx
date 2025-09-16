import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Group,
  Menu,
  MultiSelect,
  Table,
  Text,
  TextInput,
  Title,
  Tooltip,
} from "@mantine/core";
import { useDebouncedValue, useElementSize } from "@mantine/hooks";
import {
  IconCircleDashedCheck,
  IconDotsVertical,
  IconExternalLink,
  IconPlus,
  IconSearch,
  IconStar,
  IconSwitch3,
  IconTags,
} from "@tabler/icons-react";
import { useMemo, useState } from "react";
import { useLocation } from "wouter";
import { Hero } from "../../components/Hero";
import { PageWrapper } from "../../components/PageWrapper.tsx";
import { useAppContext } from "../../hooks/useAppContext.tsx";
import { useWorkspaceExecutableCounts } from "../../hooks/useWorkspaceExecutableCounts";
import { EnrichedWorkspace } from "../../types/workspace";
import { shortenPath } from "../../utils/paths";
import classes from "./Workspaces.module.css";

export function Workspaces() {
  const { workspaces, config } = useAppContext();
  const { ref: tableRef, width: tableWidth } = useElementSize();
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [debouncedSearch] = useDebouncedValue(searchQuery, 300);

  // Get executable counts for all workspaces efficiently
  const { getCountForWorkspace, isLoadingForWorkspace } =
    useWorkspaceExecutableCounts(workspaces || []);

  // Navigation hook for row clicks
  const [, setLocation] = useLocation();

  // Estimate location column width (roughly 35% of table width, minus padding and margins)
  const locationColumnWidth = Math.max(150, tableWidth * 0.35 - 60);

  // Get all unique tags from workspaces
  const allTags = useMemo(() => {
    const tagSet = new Set<string>();
    workspaces?.forEach((workspace) => {
      workspace.tags?.forEach((tag) => tagSet.add(tag));
    });
    return Array.from(tagSet).sort();
  }, [workspaces]);

  // Filter and sort workspaces
  const filteredWorkspaces = useMemo(() => {
    let filtered = workspaces || [];

    // Apply text search filter
    if (debouncedSearch.trim()) {
      const query = debouncedSearch.toLowerCase().trim();
      filtered = filtered.filter(
        (workspace) =>
          workspace.name.toLowerCase().includes(query) ||
          workspace.displayName?.toLowerCase().includes(query) ||
          workspace.description?.toLowerCase().includes(query) ||
          workspace.tags?.some((tag) => tag.toLowerCase().includes(query))
      );
    }

    // Apply tag filter
    if (selectedTags.length > 0) {
      filtered = filtered.filter((workspace) =>
        workspace.tags?.some((tag) => selectedTags.includes(tag))
      );
    }

    // Sort alphabetically by name
    return filtered.sort((a, b) => a.name.localeCompare(b.name));
  }, [workspaces, debouncedSearch, selectedTags]);

  const handleWorkspaceAction = (
    action: string,
    workspace: EnrichedWorkspace
  ) => {
    console.log(`Action: ${action} on workspace:`, workspace.name);
    // TODO: Implement actual actions
  };

  // Helper function to render executable count with loading/error states
  const renderExecutableCount = (workspaceName: string) => {
    if (isLoadingForWorkspace(workspaceName)) {
      return "loading...";
    }

    const count = getCountForWorkspace(workspaceName);
    return `${count} executable${count !== 1 ? "s" : ""}`;
  };

  // Handle row click to navigate to workspace page
  const handleRowClick = (workspaceName: string) => {
    setLocation(`/workspace/${workspaceName}`);
  };

  const rows = filteredWorkspaces.map((workspace) => (
    <Table.Tr key={workspace.name}>
      <Table.Td>
        {config != undefined && config.currentWorkspace == workspace.name && (
          <Tooltip label="Current workspace" position="right" withArrow>
            <IconCircleDashedCheck
              className={classes.currentWorkspaceIndicator}
              size={18}
            />
          </Tooltip>
        )}
      </Table.Td>
      <Table.Td className={classes.workspaceCell}>
        <Box className={classes.workspaceDetails}>
          <Text
            className={classes.workspaceNameLink}
            onClick={() => handleRowClick(workspace.name)}
          >
            {workspace.displayName || workspace.name}
          </Text>
          <Text className={classes.executableCount}>
            ({renderExecutableCount(workspace.name)})
          </Text>
        </Box>
      </Table.Td>
      <Table.Td className={classes.locationCell}>
        <Tooltip label={workspace.path} position="top" withArrow>
          <Text className={classes.pathText}>
            {shortenPath(workspace.path, locationColumnWidth)}
          </Text>
        </Tooltip>
      </Table.Td>
      <Table.Td className={classes.tagsCell}>
        <Group className={classes.tagsList} gap="xs">
          {workspace.tags?.map((tag, index) => (
            <Badge
              key={index}
              className={classes.tag}
              size="sm"
              variant="light"
            >
              {tag}
            </Badge>
          ))}
        </Group>
      </Table.Td>
      <Table.Td className={classes.actionsCell}>
        <Menu shadow="md" width={200} position="bottom-end">
          <Menu.Target>
            <ActionIcon
              variant="subtle"
              className={classes.actionButton}
              aria-label="Workspace actions"
            >
              <IconDotsVertical size={16} />
            </ActionIcon>
          </Menu.Target>

          <Menu.Dropdown>
            <Menu.Item
              leftSection={<IconSwitch3 size={14} />}
              onClick={() => handleWorkspaceAction("switch", workspace)}
            >
              Switch to workspace
            </Menu.Item>
            <Menu.Item
              leftSection={<IconExternalLink size={14} />}
              onClick={() => handleWorkspaceAction("open", workspace)}
            >
              Open in explorer
            </Menu.Item>
            <Menu.Divider />
            <Menu.Item
              leftSection={<IconStar size={14} />}
              onClick={() => handleWorkspaceAction("star", workspace)}
            >
              Star workspace
            </Menu.Item>
          </Menu.Dropdown>
        </Menu>
      </Table.Td>
    </Table.Tr>
  ));

  return (
    <>
      <PageWrapper>
        <Hero variant="split" pattern="subtle">
          <Hero.Header>
            <Title order={2}>Workspaces</Title>
            <Text c="dimmed">Organize your automation workflows</Text>
          </Hero.Header>
          <Hero.Actions>
            <Badge variant="light" size="sm" c="dimmed">
              {workspaces.length} registered
            </Badge>
            <Button leftSection={<IconPlus size={14} />} variant="filled">
              Create workspace
            </Button>
          </Hero.Actions>
        </Hero>

        <Box className={classes.filterSection}>
          <Group gap="md" className={classes.filterGroup}>
            <TextInput
              className={classes.searchInput}
              placeholder="Search workspaces..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.currentTarget.value)}
            />
            <MultiSelect
              className={classes.tagFilter}
              placeholder="Filter by tags..."
              leftSection={<IconTags size={16} />}
              data={allTags}
              value={selectedTags}
              onChange={setSelectedTags}
              clearable
              searchable
              maxDropdownHeight={200}
            />
          </Group>
        </Box>

        <Box className={classes.tableContainer}>
          <Table ref={tableRef} className={classes.table} stickyHeader>
            <Table.Thead>
              <Table.Tr>
                <Table.Th style={{ width: 40 }}></Table.Th>
                <Table.Th>Workspace</Table.Th>
                <Table.Th>Location</Table.Th>
                <Table.Th>Tags</Table.Th>
                <Table.Th style={{ width: 80 }}></Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {rows.length > 0 ? (
                rows
              ) : (
                <Table.Tr>
                  <Table.Td colSpan={5}>
                    <Box className={classes.emptyState}>
                      <IconSearch size={48} stroke={1.5} />
                      <Text className={classes.emptyStateText}>
                        {searchQuery || selectedTags.length > 0
                          ? "No workspaces match your filters"
                          : "No workspaces found"}
                      </Text>
                    </Box>
                  </Table.Td>
                </Table.Tr>
              )}
            </Table.Tbody>
          </Table>
        </Box>
      </PageWrapper>
    </>
  );
}
