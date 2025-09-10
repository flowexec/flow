import {
  Badge,
  Button,
  Card,
  Divider,
  Grid,
  Group,
  Stack,
  Text,
  ThemeIcon,
  Title,
  Tooltip,
} from "@mantine/core";
import {
  IconExternalLink,
  IconFilter,
  IconFolder,
  IconInfoCircle,
  IconTag,
} from "@tabler/icons-react";
import { openPath } from "@tauri-apps/plugin-opener";
import { MarkdownRenderer } from "../../components/MarkdownRenderer";
import { useSettings } from "../../hooks/useSettings";
import { EnrichedWorkspace } from "../../types/workspace";

interface WorkspaceProps {
  workspace: EnrichedWorkspace | null;
}

export function Workspace({ workspace }: WorkspaceProps) {
  const { settings } = useSettings();

  if (!workspace) {
    return null;
  }

  const onOpenDir = async () => {
    try {
      await openPath(workspace.path, settings.workspaceApp || undefined);
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <Stack gap="lg">
      <Card withBorder>
        <Stack gap="md">
          <Group justify="space-between" align="flex-start">
            <Stack gap="xs">
              <Group gap="sm" align="center">
                <ThemeIcon variant="light" size="lg">
                  <IconFolder size={20} />
                </ThemeIcon>
                <div>
                  <Title order={2}>
                    {workspace.displayName || workspace.name}
                  </Title>
                  <Text size="sm" c="dimmed">
                    Workspace
                  </Text>
                </div>
              </Group>

              {workspace.name && (
                <Tooltip label={`Registered at ${workspace.path}`}>
                  <Badge variant="light" color="gray">
                    <Group gap={4}>
                      <IconInfoCircle size={12} />
                      {workspace.name}
                    </Group>
                  </Badge>
                </Tooltip>
              )}
            </Stack>

            <Group gap="sm">
              {onOpenDir && (
                <Button
                  variant="light"
                  leftSection={<IconExternalLink size={16} />}
                  onClick={onOpenDir}
                  size="md"
                >
                  Open
                </Button>
              )}
            </Group>
          </Group>

          {workspace.fullDescription && (
            <>
              <Divider />
              <MarkdownRenderer>{workspace.fullDescription}</MarkdownRenderer>
            </>
          )}
        </Stack>
      </Card>

      <Grid>
        {workspace.tags && workspace.tags.length > 0 && (
          <Grid.Col span={6}>
            <Stack gap="md">
              <Card withBorder>
                <Stack gap="sm">
                  <Title order={4}>
                    <Group gap="xs">
                      <IconTag size={16} />
                      Tags
                    </Group>
                  </Title>
                  <Group gap="xs">
                    {workspace.tags.map((tag, index) => (
                      <Badge key={index} variant="dot">
                        {tag}
                      </Badge>
                    ))}
                  </Group>
                </Stack>
              </Card>
            </Stack>
          </Grid.Col>
        )}

        {workspace.executables && (
          <Grid.Col span={6}>
            <Stack gap="md">
              <Card withBorder>
                <Stack gap="sm">
                  <Title order={4}>
                    <Group gap="xs">
                      <IconFilter size={16} />
                      Executable Filters
                    </Group>
                  </Title>
                  <Stack gap="xs">
                    {workspace.executables.included &&
                      workspace.executables.included.length > 0 && (
                        <div>
                          <Text size="sm" fw={500}>
                            Included:
                          </Text>
                          <Group gap="xs">
                            {workspace.executables.included.map(
                              (path, index) => (
                                <Badge
                                  key={index}
                                  variant="light"
                                  color="green"
                                >
                                  {path}
                                </Badge>
                              ),
                            )}
                          </Group>
                        </div>
                      )}
                    {workspace.executables.excluded &&
                      workspace.executables.excluded.length > 0 && (
                        <div>
                          <Text size="sm" fw={500}>
                            Excluded:
                          </Text>
                          <Group gap="xs">
                            {workspace.executables.excluded.map(
                              (path, index) => (
                                <Badge key={index} variant="light" color="red">
                                  {path}
                                </Badge>
                              ),
                            )}
                          </Group>
                        </div>
                      )}
                  </Stack>
                </Stack>
              </Card>
            </Stack>
          </Grid.Col>
        )}
      </Grid>
    </Stack>
  );
}
