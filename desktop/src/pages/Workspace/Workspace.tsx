import {
  Badge,
  Button,
  Card,
  Grid,
  Group,
  List,
  LoadingOverlay,
  Stack,
  Table,
  Text,
  ThemeIcon,
  Title,
  Tooltip,
} from "@mantine/core";
import {
  IconArrowRight,
  IconArrowsJoin,
  IconCurrencyDollar,
  IconExternalLink,
  IconFilter,
  IconFolder,
  IconInfoCircle,
  IconTag,
} from "@tabler/icons-react";
import { useParams } from "wouter";
import { Hero } from "../../components/Hero";
import { MarkdownRenderer } from "../../components/MarkdownRenderer";
import { PageWrapper } from "../../components/PageWrapper.tsx";
import { useWorkspace } from "../../hooks/useWorkspace";
import { useWorkspaceExecutableCount } from "../../hooks/useWorkspaceExecutableCounts.ts";
import { stringToColor } from "../../utils/colors.ts";

export function Workspace() {
  const { workspaceName } = useParams();
  const { workspace, workspaceError, isWorkspaceLoading } = useWorkspace(
    workspaceName || ""
  );

  const { count: executableCount, isLoading: isExecutableCountLoading } =
    useWorkspaceExecutableCount(workspace?.name);

  return (
    <PageWrapper>
      {isWorkspaceLoading && (
        <LoadingOverlay
          visible={isWorkspaceLoading}
          zIndex={1000}
          overlayProps={{ radius: "sm", blur: 2 }}
        />
      )}
      {workspaceError && <Text c="red">Error: {workspaceError.message}</Text>}
      {workspace ? (
        <>
          <Hero variant="split" pattern="subtle">
            <Hero.Header>
              <Group gap="xs">
                <ThemeIcon variant="light" size="lg">
                  <IconFolder size={16} />
                </ThemeIcon>
                <Title order={2}>
                  {workspace.displayName || workspace.name}
                </Title>
              </Group>
              <Text
                size="xs"
                pl="calc(34px + var(--mantine-spacing-xs))"
              >
                {isExecutableCountLoading
                  ? "loading..."
                  : `${executableCount} executable${executableCount !== 1 ? "s" : ""}`}
              </Text>
            </Hero.Header>
            <Hero.Actions>
              {workspace.name && (
                <Tooltip label={`Registered at ${workspace.path}`}>
                  <Badge size="sm" c="dimmed" color="tertiary" variant="white">
                    <Group gap={4}>
                      <IconInfoCircle size={12} />
                      {workspace.name}
                    </Group>
                  </Badge>
                </Tooltip>
              )}
              <Group gap="sm">
                <Button
                  leftSection={<IconExternalLink size={16} />}
                  size="compact-xs"
                  color="tertiary"
                >
                  Open
                </Button>
              </Group>
            </Hero.Actions>
          </Hero>

          <Stack gap="lg">
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
                            <Badge
                              key={index}
                              color={stringToColor(tag)}
                              size="sm"
                              autoContrast
                            >
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
                                    )
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
                                      <Badge
                                        key={index}
                                        variant="light"
                                        color="red"
                                      >
                                        {path}
                                      </Badge>
                                    )
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

              {workspace.verbAliases &&
                Object.keys(workspace.verbAliases).length > 0 && (
                  <Grid.Col span={6}>
                    <Stack gap="md">
                      <Card withBorder>
                        <Stack gap="sm">
                          <Title order={4}>
                            <Group gap="xs">
                              <IconArrowsJoin size={16} />
                              Verb Aliases
                            </Group>
                          </Title>
                          <Stack gap="xs">
                            <Table withRowBorders={false}>
                              {Object.entries(workspace.verbAliases).map(
                                ([mainVerb, aliases]) => (
                                  <Table.Tr key={mainVerb}>
                                    <Table.Td>{mainVerb}</Table.Td>
                                    <Table.Td>
                                      <IconArrowRight size={12} />{" "}
                                      {aliases.join(", ")}
                                    </Table.Td>
                                  </Table.Tr>
                                )
                              )}
                            </Table>
                          </Stack>
                        </Stack>
                      </Card>
                    </Stack>
                  </Grid.Col>
                )}

              {workspace.envFiles && workspace.envFiles.length > 0 && (
                <Grid.Col span={6}>
                  <Stack gap="md">
                    <Card withBorder>
                      <Stack gap="sm">
                        <Title order={4}>
                          <Group gap="xs">
                            <IconCurrencyDollar size={16} />
                            Env Files
                          </Group>
                        </Title>
                        <Stack gap="xs">
                          <List>
                            {workspace.envFiles.map((envFile, index) => (
                              <List.Item key={index}>{envFile}</List.Item>
                            ))}
                          </List>
                        </Stack>
                      </Stack>
                    </Card>
                  </Stack>
                </Grid.Col>
              )}
            </Grid>

            {workspace.fullDescription && (
              <>
                <MarkdownRenderer>{workspace.fullDescription}</MarkdownRenderer>
              </>
            )}
          </Stack>
        </>
      ) : (
        <Text c="red">Error: Workspace not found</Text>
      )}
    </PageWrapper>
  );
}
