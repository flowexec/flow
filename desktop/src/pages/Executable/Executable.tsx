import {
  Badge,
  Button,
  Card,
  Code,
  Drawer,
  Grid,
  Group,
  Stack,
  Text,
  ThemeIcon,
  Title,
  Tooltip,
  LoadingOverlay,
  Alert, ButtonGroup,
} from "@mantine/core";
import {
  IconArrowsSplit,
  IconClock,
  IconExternalLink,
  IconEye,
  IconFile,
  IconLabel,
  IconPlayerPlay,
  IconRoute,
  IconTag,
  IconTemplate,
  IconTerminal,
} from "@tabler/icons-react";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import { openPath } from "@tauri-apps/plugin-opener";
import { useEffect, useState } from "react";
import { useParams } from "wouter";
import { MarkdownRenderer } from "../../components/MarkdownRenderer";
import { useNotifier } from "../../hooks/useNotifier";
import { useSettings } from "../../hooks/useSettings";
import { useExecutable } from "../../hooks/useExecutable";
import { PageWrapper } from "../../components/PageWrapper.tsx";
import { Hero } from "../../components/Hero";
import { EnrichedExecutable } from "../../types/executable";
import { NotificationType } from "../../types/notification";
import { LogLine, LogViewer } from "../Logs/LogViewer";
import { ExecutableEnvironmentDetails } from "./ExecutableEnvironmentDetails";
import { ExecutableTypeDetails } from "./ExecutableTypeDetails";
import { ExecutionForm, ExecutionFormData } from "./ExecutionForm";
import { stringToColor } from "../../utils/colors.ts";

export function Executable() {
  const params = useParams();
  const executableId = decodeURIComponent(params.executableId || "");
  const { executable, executableError, isExecutableLoading } =
    useExecutable(executableId);

  // Local UI state and helpers formerly in ExecutableView
  const { settings } = useSettings();
  const { setNotification } = useNotifier();
  const [output, setOutput] = useState<LogLine[]>([]);
  const [formOpened, setFormOpened] = useState(false);

  useEffect(() => {
    let unlistenOutput: (() => void) | undefined;
    let unlistenComplete: (() => void) | undefined;

    const setupListeners = async () => {
      unlistenOutput = await listen("command-output", (event) => {
        const payload = event.payload as LogLine;
        setOutput((prev) => [...prev, payload]);
      });

      unlistenComplete = await listen("command-complete", (event) => {
        const payload = event.payload as {
          success: boolean;
          exit_code: number | null;
        };
        setNotification({
          title: payload.success ? "Execution completed" : "Execution failed",
          message: payload.success
            ? "Execution completed successfully"
            : "Execution failed",
          type: payload.success
            ? NotificationType.Success
            : NotificationType.Error,
        });
      });
    };

    void setupListeners();

    return () => {
      if (unlistenOutput) unlistenOutput();
      if (unlistenComplete) unlistenComplete();
    };
  }, [setNotification]);

  const onOpenFile = async () => {
    if (!executable) return;
    try {
      await openPath(executable.flowfile, settings.executableApp || undefined);
    } catch (error) {
      console.error(error);
    }
  };

  const executeWithData = async (formData: ExecutionFormData) => {
    if (!executable) return;
    try {
      setOutput([]);

      setNotification({
        title: "Execution started",
        message: `Execution of ${executable.ref} started`,
        type: NotificationType.Success,
        autoClose: true,
        autoCloseDelay: 5000,
      });

      const argsArray = formData.args.trim()
        ? formData.args.trim().split(/\s+/)
        : [];

      const invokeParams: {
        verb: string;
        executableId: string;
        args: string[];
        params?: Record<string, string>;
      } = {
        verb: executable.verb,
        executableId: executable.id,
        args: argsArray,
      };

      if (Object.keys(formData.params).length > 0) {
        invokeParams.params = formData.params;
      }

      await invoke("execute", invokeParams);
    } catch (error) {
      console.error(error);
      setNotification({
        title: "Execution failed",
        message: `Execution of ${executable.ref} failed`,
        type: NotificationType.Error,
        autoClose: true,
        autoCloseDelay: 5000,
      });
    }
  };

  const onExecute = async () => {
    if (!executable) return;
    const hasPromptParams = executable.exec?.params?.some((param) => param.prompt);
    const hasArgs = executable.exec?.args && executable.exec.args.length > 0;

    if (hasPromptParams || hasArgs) {
      setFormOpened(true);
      return;
    }

    await executeWithData({ params: {}, args: "" });
  };

  const typeInfo = executable && getExecutableTypeInfo(executable);

  return (
    <PageWrapper>
      {isExecutableLoading && (
        <LoadingOverlay
          visible={isExecutableLoading}
          zIndex={1000}
          overlayProps={{ radius: "sm", blur: 2 }}
        />
      )}
      {executableError && <Alert variant="light" color="red.5">Error: {executableError.message}</Alert>}
      {executable ? (
        <Stack gap="sm">
          <Hero variant="left" pattern="subtle">
            <Hero.Header>
              <Group gap="xs">
                <ThemeIcon variant="light" size="lg">
                  {typeInfo && <typeInfo.icon size={16} />}
                </ThemeIcon>
                <Title order={2}>{executable.ref}</Title>
              </Group>
              {typeInfo?.description && (
                <Text size="xs" c="dimmed" pl="calc(34px + var(--mantine-spacing-xs))">{typeInfo.description}</Text>
              )}
            </Hero.Header>
            <Hero.Actions>
              <Badge variant="light" color={getVisibilityColor(executable.visibility)}>
                <Group gap={4}>
                  <IconEye size={12} />
                  {executable.visibility || "public"}
                </Group>
              </Badge>
              {executable.timeout && (
                <Badge variant="light" color="gray">
                  <Group gap={4}>
                    <IconClock size={12} />
                    {executable.timeout}
                  </Group>
                </Badge>
              )}
              <Tooltip label={`Defined in ${executable.flowfile}`}>
                <Badge size="sm" color="tertiary" variant="white">
                  <Group gap={4}>
                    <IconFile size={12} />
                    {executable.flowfile.split("/").pop() || executable.flowfile}
                  </Group>
                </Badge>
              </Tooltip>
              <ButtonGroup>
                <Button onClick={onOpenFile} leftSection={<IconExternalLink size={16} />} size="compact-xs" color="tertiary" variant="white">
                  Edit
                </Button>
                <Button onClick={onExecute} leftSection={<IconPlayerPlay size={16} />} size="compact-xs" color="tertiary">
                  Execute
                </Button>
              </ButtonGroup>
            </Hero.Actions>
          </Hero>

          {executable.description && (
            <>
              <MarkdownRenderer>{executable.description}</MarkdownRenderer>
            </>
          )}

          <Grid>
            {executable.aliases && executable.aliases.length > 0 && (
              <Grid.Col span={6}>
                <Card withBorder>
                  <Stack gap="sm">
                    <Title order={4}>
                      <Group gap="xs">
                        <IconLabel size={16} />
                        Aliases
                      </Group>
                    </Title>
                    <Group gap="xs">
                      {executable.aliases.map((alias, index) => (
                        <Code key={index}>{alias}</Code>
                      ))}
                    </Group>
                  </Stack>
                </Card>
              </Grid.Col>
            )}

            {executable.verbAliases && executable.verbAliases.length > 0 && (
              <Grid.Col span={6}>
                <Card withBorder>
                  <Stack gap="sm">
                    <Title order={4}>
                      <Group gap="xs">
                        <IconLabel size={16} />
                        Verb Aliases
                      </Group>
                    </Title>
                    <Group gap="xs">
                      {executable.verbAliases.map((alias, index) => (
                        <Code key={index}>{String(alias)}</Code>
                      ))}
                    </Group>
                  </Stack>
                </Card>
              </Grid.Col>
            )}

            {executable.tags && executable.tags.length > 0 && (
              <Grid.Col span={6}>
                <Card withBorder>
                  <Stack gap="sm">
                    <Title order={4}>
                      <Group gap="xs">
                        <IconTag size={16} />
                        Tags
                      </Group>
                    </Title>
                    <Group gap="xs">
                      {executable.tags.map((tag, index) => (
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
              </Grid.Col>
            )}
          </Grid>

          <ExecutableEnvironmentDetails executable={executable} />
          <ExecutableTypeDetails executable={executable} />

          {output.length > 0 && (
            <Drawer
              opened={true}
              onClose={() => setOutput([])}
              title={<Text size="sm">Execution Output</Text>}
              size="33%"
              position="bottom"
            >
              <LogViewer logs={output} formatted={true} fontSize={12} />
            </Drawer>
          )}

          {formOpened && (
            <ExecutionForm
              opened={formOpened}
              onClose={() => setFormOpened(false)}
              onSubmit={executeWithData}
              executable={executable}
            />
          )}
        </Stack>
      ) : (
        <Alert variant="light" color="red.5">Error: Executable not found</Alert>
      )}
    </PageWrapper>
  );
}

function getExecutableTypeInfo(executable: EnrichedExecutable) {
  if (executable.exec)
    return {
      type: "exec",
      icon: IconTerminal,
      description: "Command execution",
    };
  if (executable.serial)
    return {
      type: "serial",
      icon: IconRoute,
      description: "Sequential execution",
    };
  if (executable.parallel)
    return {
      type: "parallel",
      icon: IconArrowsSplit,
      description: "Parallel execution",
    };
  if (executable.launch)
    return {
      type: "launch",
      icon: IconExternalLink,
      description: "Launch application/URI",
    };
  if (executable.request)
    return {
      type: "request",
      icon: IconExternalLink,
      description: "HTTP request",
    };
  if (executable.render)
    return {
      type: "render",
      icon: IconTemplate,
      description: "Render template",
    };
  return { type: "unknown", icon: IconTerminal, description: "Unknown type" };
}

function getVisibilityColor(visibility?: string) {
  switch (visibility) {
    case "public":
      return "green.3";
    case "private":
      return "blue.3";
    case "internal":
      return "orange.3";
    case "hidden":
      return "red.3";
    default:
      return "gray.3";
  }
}
