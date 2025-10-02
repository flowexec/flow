import {
  Badge,
  Box,
  Button,
  Drawer,
  Group,
  MultiSelect,
  Select,
  Stack,
  Text,
  TextInput,
} from "@mantine/core";
import {
  IconAbc,
  IconBraces,
  IconCategory,
  IconEye,
  IconFilter,
  IconFolder,
  IconSearch,
  IconTags,
} from "@tabler/icons-react";
import classes from "./Executables.module.css";
import { CommonVisibility } from "../../types/generated/flowfile.ts";

export interface FilterState {
  workspace: string;
  tags: string[];
  namespace: string;
  verb: string;
  search: string;
  type: string;
  visibility: CommonVisibility | '';
}

export interface WorkspaceOption {
  label: string;
  value: string;
}

export interface FilterDrawerProps {
  opened: boolean;
  onClose: () => void;
  filterState: FilterState;
  onFilterChange: (updates: Partial<FilterState>) => void;
  onClearAll: () => void;
  // Options for dropdowns
  workspaceOptions: WorkspaceOption[];
  namespaceOptions: string[];
  tagOptions: string[];
  verbOptions: string[];
}

export function FilterButton({
 activeCount,
 onClick
}: {
  activeCount: number;
  onClick: () => void;
}) {
  return (
    <Button
      size="xs"
      variant={activeCount > 0 ? "filled" : "light"}
      leftSection={<IconFilter size={16} />}
      rightSection={activeCount > 0 ? (
        <Badge size="xs" color="var(--mantine-color-info-0)" variant="white">
          {activeCount}
        </Badge>
      ) : undefined}
      onClick={onClick}
    >
      Filters
    </Button>
  );
}

export function FilterDrawer({
 opened,
 onClose,
 filterState,
 onFilterChange,
 onClearAll,
 workspaceOptions,
 namespaceOptions,
 tagOptions,
 verbOptions,
}: FilterDrawerProps) {
  const handleFilterChange = (field: keyof FilterState) => (value: any) => {
    onFilterChange({ [field]: value });
  };

  return (
    <Drawer
      opened={opened}
      onClose={onClose}
      position="right"
      title="Filter Executables"
      size="sm"
      classNames={{title: classes.filterTitle}}
    >
      <Stack justify="space-between" gap="xl">
        <Box>
          <Stack mt="sm" gap="xs">
            <TextInput
              size="xs"
              placeholder="Search"
              leftSection={<IconSearch size={16} />}
              value={filterState.search}
              onChange={(e) => handleFilterChange('search')(e.currentTarget.value)}
            />
            <MultiSelect
              size="xs"
              placeholder="Tags"
              leftSection={<IconTags size={16} />}
              data={tagOptions}
              value={filterState.tags}
              onChange={handleFilterChange('tags')}
              clearable
              searchable
            />
          </Stack>

          <Stack mt="xl" gap="xs">
            <Text size='xs' className={classes.filterLabel}>Scope</Text>
            <Select
              size="xs"
              placeholder="Workspace"
              leftSection={<IconFolder size={16} />}
              data={workspaceOptions}
              value={filterState.workspace}
              onChange={handleFilterChange('workspace')}
              clearable
              searchable
            />
            <Select
              size="xs"
              placeholder="Namespace"
              leftSection={<IconCategory size={16} />}
              data={namespaceOptions}
              value={filterState.namespace}
              onChange={handleFilterChange('namespace')}
              clearable
              searchable
            />
          </Stack>

          <Stack mt="xl" gap="xs">
            <Text size='xs' className={classes.filterLabel}>Properties</Text>
            <Select
              size="xs"
              placeholder="Verb"
              leftSection={<IconAbc size={16} />}
              data={verbOptions}
              value={filterState.verb}
              onChange={handleFilterChange('verb')}
              clearable
              searchable
            />
            <Select
              size="xs"
              placeholder="Type"
              leftSection={<IconBraces size={16} />}
              data={["command", "launch", "render", "request",  "serial", "parallel"]}
              value={filterState.type}
              onChange={handleFilterChange('type')}
              clearable
            />
            <Select
              size="xs"
              placeholder="Visibility"
              leftSection={<IconEye size={16} />}
              data={["public", "private", "internal"]}
              value={filterState.visibility}
              onChange={handleFilterChange('visibility')}
              clearable
            />
          </Stack>
        </Box>

        <Group mt="xl" gap="sm" justify="flex-end">
          <Button size="xs" variant="light" onClick={onClearAll}>
            Clear all
          </Button>
          <Button size="xs" onClick={onClose}>
            Close
          </Button>
        </Group>
      </Stack>
    </Drawer>
  );
}