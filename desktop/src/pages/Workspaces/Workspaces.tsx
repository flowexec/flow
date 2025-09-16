import {
  ActionIcon,
  Badge,
  Box,
  Button,
  Group,
  Table,
  Text,
  Title,
  Tooltip,
} from "@mantine/core";
import { useElementSize } from "@mantine/hooks";
import {
  IconCircleDashedCheck,
  IconExternalLink,
  IconPlus,
  IconStar,
  IconSwitch3,
} from "@tabler/icons-react";
import { Hero } from "../../components/Hero";
import { useAppContext } from "../../hooks/useAppContext.tsx";
import { shortenPath } from "../../utils/paths";

export function Workspaces() {
  // const { settings } = useSettings();
  const { workspaces, config } = useAppContext();
  const { ref: tableRef, width: tableWidth } = useElementSize();

  // Estimate location column width (roughly 35% of table width, minus padding and margins)
  const locationColumnWidth = Math.max(150, tableWidth * 0.35 - 60);

  const rows = workspaces.map((workspace) => (
    <Table.Tr key={workspace.name}>
      <Table.Td>
        {config != undefined && config.currentWorkspace == workspace.name && (
          <IconCircleDashedCheck />
        )}
      </Table.Td>
      <Table.Td>
        <Box>
          {workspace.displayName}
          <Text size="xs">(X executables)</Text>
        </Box>
      </Table.Td>
      <Table.Td>
        <Tooltip label={workspace.path} position="top" withArrow>
          <Text
            style={{
              cursor: "default",
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {shortenPath(workspace.path, locationColumnWidth)}
          </Text>
        </Tooltip>
      </Table.Td>
      <Table.Td>{workspace.tags?.map((tag) => <Badge>{tag}</Badge>)}</Table.Td>
      <Table.Td>
        <Group>
          <ActionIcon variant="filled" aria-label="Settings">
            <IconSwitch3 />
          </ActionIcon>
          <ActionIcon variant="filled" aria-label="Settings">
            <IconExternalLink />
          </ActionIcon>
          <ActionIcon variant="filled" aria-label="Settings">
            <IconStar />
          </ActionIcon>
        </Group>
      </Table.Td>
    </Table.Tr>
  ));

  return (
    <>
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

      <Table ref={tableRef}>
        <Table.Thead>
          <Table.Tr>
            <Table.Th></Table.Th>
            <Table.Th>Workspace</Table.Th>
            <Table.Th>Location</Table.Th>
            <Table.Th>Tags</Table.Th>
            <Table.Th></Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>{rows}</Table.Tbody>
      </Table>
    </>
  );
}
