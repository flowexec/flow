import { ActionIcon, Group } from "@mantine/core";
import { IconPlus, IconRefresh } from "@tabler/icons-react";

interface ActionButtonsProps {
  onCreateWorkspace: () => void;
  onRefreshWorkspaces: () => void;
}

export function ActionButtons({
  onCreateWorkspace,
  onRefreshWorkspaces,
}: ActionButtonsProps) {
  return (
    <Group>
      <ActionIcon
        color="palette"
        onClick={onCreateWorkspace}
        title="Create workspace"
        size="md"
      >
        <IconPlus />
      </ActionIcon>
      <ActionIcon
        color="palette"
        onClick={onRefreshWorkspaces}
        title="Refresh workspaces"
        size="md"
      >
        <IconRefresh />
      </ActionIcon>
    </Group>
  );
}
