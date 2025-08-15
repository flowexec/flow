import {
  Badge,
  Button,
  Card,
  Code,
  Divider,
  Drawer,
  Grid,
  Group,
  Stack,
  Text,
  ThemeIcon,
  Title,
  Tooltip,
} from "@mantine/core";
import {
  IconClock,
  IconExternalLink,
  IconEye,
  IconFile,
  IconLabel,
  IconPlayerPlay,
  IconTag,
  IconTerminal,
} from "@tabler/icons-react";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import { openPath } from "@tauri-apps/plugin-opener";
import { useEffect, useState } from "react";
import { MarkdownRenderer } from "../../components/MarkdownRenderer";
import { useNotifier } from "../../hooks/useNotifier";
import { useSettings } from "../../hooks/useSettings";
import { EnrichedExecutable } from "../../types/executable";
import { NotificationType } from "../../types/notification";
import { LogLine, LogViewer } from "../Logs/LogViewer";
import { ExecutableEnvironmentDetails } from "./ExecutableEnvironmentDetails";
import { ExecutableTypeDetails } from "./ExecutableTypeDetails";
import { ExecutionForm, ExecutionFormData } from "./ExecutionForm";

export type ExecutableProps = {
  executable: EnrichedExecutable;
};

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
      icon: IconTerminal,
      description: "Sequential execution",
    };
  if (executable.parallel)
    return {
      type: "parallel",
      icon: IconTerminal,
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
      icon: IconTerminal,
      description: "Render template",
    };
  return { type: "unknown", icon: IconTerminal, description: "Unknown type" };
}

function getVisibilityColor(visibility?: string) {
  switch (visibility) {
    case "public":
      return "green";
    case "private":
      return "blue";
    case "internal":
      return "orange";
    case "hidden":
      return "red";
    default:
      return "gray";
  }
}

export function Executable({ executable }: ExecutableProps) {
  const typeInfo = getExecutableTypeInfo(executable);
  const { settings } = useSettings();
  const { setNotification } = useNotifier();
  const [output, setOutput] = useState<LogLine[]>([]);
  const [formOpened, setFormOpened] = useState(false);

  useEffect(() => {
    let unlistenOutput: (() => void) | undefined;
    let unlistenComplete: (() => void) | undefined;

    const setupListeners = async () => {
      console.log("Setting up listeners for executable:", executable.ref);
      unlistenOutput = await listen("command-output", (event) => {
        const payload = event.payload as LogLine;
        console.log("Received output:", payload);
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

    setupListeners();

    return () => {
      if (unlistenOutput) {
        unlistenOutput();
      }
      if (unlistenComplete) {
        unlistenComplete();
      }
    };
  }, [setNotification]);

  const onOpenFile = async () => {
    try {
      console.log(executable.flowfile);
      await openPath(executable.flowfile, settings.executableApp || undefined);
    } catch (error) {
      console.error(error);
    }
  };

  const onExecute = async () => {
    const hasPromptParams = executable.exec?.params?.some(
      (param) => param.prompt
    );
    const hasArgs = executable.exec?.args && executable.exec.args.length > 0;

    if (hasPromptParams || hasArgs) {
      setFormOpened(true);
      return;
    }

    await executeWithData({ params: {}, args: "" });
  };

  const executeWithData = async (formData: ExecutionFormData) => {
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

      const invokeParams: any = {
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

  return (
    <Stack gap="lg">
      <Card withBorder>
        <Stack gap="md">
          <Group justify="space-between" align="flex-start">
            <Stack gap="xs">
              <Group gap="sm" align="center">
                <ThemeIcon variant="light" size="lg">
                  <typeInfo.icon size={20} />
                </ThemeIcon>
                <div>
                  <Title order={2}>{executable.ref}</Title>
                  <Text size="sm" c="dimmed">
                    {typeInfo.description}
                  </Text>
                </div>
              </Group>

              <Group gap="xs">
                <Tooltip label={`Defined in ${executable.flowfile}`}>
                  <Badge variant="light" size="sm">
                    <Group gap={4}>
                      <IconFile size={12} />
                      {executable.flowfile.split("/").pop() ||
                        executable.flowfile}
                    </Group>
                  </Badge>
                </Tooltip>
                <Badge
                  variant="light"
                  color={getVisibilityColor(executable.visibility)}
                >
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
              </Group>
            </Stack>

            <Group gap="sm">
              <Button
                leftSection={<IconPlayerPlay size={16} />}
                onClick={onExecute}
                size="md"
              >
                Execute
              </Button>
              <Button
                variant="light"
                leftSection={<IconExternalLink size={16} />}
                onClick={onOpenFile}
                size="md"
              >
                Edit
              </Button>
            </Group>
          </Group>

          {executable.description && (
            <>
              <Divider />
              <MarkdownRenderer>{executable.description}</MarkdownRenderer>
            </>
          )}
        </Stack>
      </Card>

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
                    <Badge key={index} variant="dot">
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
  );
}
