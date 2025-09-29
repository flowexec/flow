import { Badge, Card, Code, Divider, Group, Stack, Title } from "@mantine/core";
import { IconHash } from "@tabler/icons-react";
import { useSettings } from "../../../hooks/useSettings";
import { EnrichedExecutable } from "../../../types/executable";
import { CodeHighlighter } from "../../../components/CodeHighlighter";

export type ExecutableExecDetailsProps = {
  executable: EnrichedExecutable;
};

export function ExecutableExecDetails({
  executable,
}: ExecutableExecDetailsProps) {
  const { settings } = useSettings();

  return (
    <Card withBorder shadow="sm" padding="md">
      <Stack gap="sm">
        <Group justify="space-between" align="center">
          <Title order={4}>
            <Group gap="xs">
              <IconHash size={16} />
              Execution Configuration
            </Group>
          </Title>
          <Group gap="xs">
            {executable.exec?.cmd && (
              <Badge variant="light" color="green.5">cmd</Badge>
            )}
            {executable.exec?.file && (
              <Badge variant="light" color="blue.5">file</Badge>
            )}
            {executable.exec?.logMode && (
              <Badge variant="light" color="gray.5">format: {executable.exec?.logMode}</Badge>
            )}
          </Group>
        </Group>

        {executable.exec?.cmd && (
          <>
            <Divider label="Command" labelPosition="left" my="xs" />
            <CodeHighlighter theme={settings.theme} copyButton={false}>
              {executable.exec.cmd}
            </CodeHighlighter>
          </>
        )}

        {executable.exec?.file && (
          <>
            <Divider label="File" labelPosition="left" my="xs" />
            <Code>{executable.exec.file}</Code>
          </>
        )}

        {executable.exec?.dir && (
          <>
            <Divider label="Working directory" labelPosition="left" my="xs" />
            <Code>{executable.exec.dir}</Code>
          </>
        )}
      </Stack>
    </Card>
  );
}
