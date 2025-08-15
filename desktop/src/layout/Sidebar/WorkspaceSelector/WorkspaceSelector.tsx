import { ComboboxItem, Group, OptionsFilter, Select } from "@mantine/core";
import { useConfig } from "../../../hooks/useConfig";
import { useNotifier } from "../../../hooks/useNotifier";
import { NotificationType } from "../../../types/notification";
import {useAppContext} from "../../../hooks/useAppContext.tsx";

const filter: OptionsFilter = ({ options, search }) => {
  const filtered = (options as ComboboxItem[]).filter((option) =>
    option.label.toLowerCase().trim().includes(search.toLowerCase().trim())
  );

  filtered.sort((a, b) => a.label.localeCompare(b.label));
  return filtered;
};

export function WorkspaceSelector() {
  const { selectedWorkspace, setSelectedWorkspace, workspaces, config } = useAppContext()
  const { updateCurrentWorkspace } = useConfig();
  const { setNotification } = useNotifier();

  const handleWorkspaceChange = async (workspaceName: string) => {
    setSelectedWorkspace(workspaceName);

    if (config?.workspaceMode === 'dynamic') {
      try {
        await updateCurrentWorkspace(workspaceName);
        setNotification({
          type: NotificationType.Success,
          title: 'Workspace switched',
          message: `Switched to workspace: ${workspaceName}`,
          autoClose: true,
        });
      } catch (error) {
        setNotification({
          type: NotificationType.Error,
          title: 'Error switching workspace',
          message: error instanceof Error ? error.message : 'An unknown error occurred',
          autoClose: true,
        });
      }
    }
  };

  return (
    <>
      <Group gap="xs" align="center" wrap="nowrap">
        <Select
          value={selectedWorkspace}
          onChange={(value) => {
            if (value && workspaces.find((w) => w.name === value)) {
              handleWorkspaceChange(value);
            }
          }}
          data={workspaces.map((workspace) => ({
            value: workspace.name,
            label: workspace.displayName || workspace.name,
          }))}
          filter={filter}
          placeholder="No workspace selected"
          searchable
          size="sm"
          styles={{
            root: {
              flex: 1,
            },
            input: {
              backgroundColor: "var(--mantine-color-dark-6)",
              borderColor: "var(--mantine-color-dark-4)",
              color: "var(--mantine-color-white)",
            },
            dropdown: {
              backgroundColor: "var(--mantine-color-dark-6)",
              borderColor: "var(--mantine-color-dark-4)",
            },
            option: {
              color: "var(--mantine-color-white)",
              "&[dataSelected]": {
                backgroundColor: "var(--mantine-color-dark-5)",
              },
              "&[dataHovered]": {
                backgroundColor: "var(--mantine-color-dark-5)",
              },
            },
          }}
        />
      </Group>
    </>
  );
}
