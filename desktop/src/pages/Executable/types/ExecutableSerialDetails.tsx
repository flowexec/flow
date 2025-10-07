import { Anchor, Badge, Card, Divider, Group, Stack, Text, Timeline, Title } from "@mantine/core";
import { IconArrowRight } from "@tabler/icons-react";
import { useSettings } from "../../../hooks/useSettings";
import { EnrichedExecutable } from "../../../types/executable";
import { CodeHighlighter } from "../../../components/CodeHighlighter";
import { Link } from "wouter";

export type ExecutableSerialDetailsProps = {
  executable: EnrichedExecutable;
};

export function ExecutableSerialDetails({
  executable,
}: ExecutableSerialDetailsProps) {
  const { settings } = useSettings();

  const execs = executable.serial?.execs || [];

  return (
    <Card withBorder shadow="sm" padding="md">
      <Stack gap="sm">
        <Group justify="space-between">
          <Title order={4}>
            <Group gap="xs">
              <IconArrowRight size={16} />
              Serial Configuration
            </Group>
          </Title>
          <Group gap="xs">
            {executable.serial?.failFast && (
              <Badge variant="light" color={executable.serial?.failFast ? "red.5" : "green.5"}>
                Fail fast: {executable.serial?.failFast ? "on" : "off"}
              </Badge>
            )}
            {execs.length > 0 && (
              <Badge variant="light" color="gray.5">Execs: {execs.length}</Badge>
            )}
          </Group>
        </Group>

        {execs.length > 0 ? (
          <>
            <Divider label="Execution order" labelPosition="left" my="xs" />
            <Timeline active={execs.length} bulletSize={20} lineWidth={2}>
              {execs.map((exec, index) => (
                <Timeline.Item
                  key={index}
                  title={
                    <Group gap={6} align="center">
                      <Text fw={600}>#{index + 1}</Text>
                      {exec.ref ? (
                        <>
                          <Text>ref: </Text>
                          <Anchor component={Link} to={`/executable/${encodeURIComponent(exec.ref)}`}>
                            {exec.ref}
                          </Anchor>
                        </>
                      ) : (
                        <Text>cmd</Text>
                      )}
                    </Group>
                  }
                >
                  <Stack gap={6}>
                    <Group gap={6}>
                      {exec.retries !== undefined && exec.retries > 0 && (
                        <Badge size="sm" variant="dot" color="orange">retries: {exec.retries}</Badge>
                      )}
                      {exec.reviewRequired && (
                        <Badge size="sm" variant="light" color="violet">review required</Badge>
                      )}
                      {exec.args && exec.args.length > 0 && (
                        <Badge size="sm" variant="light" color="gray">args: {exec.args.length}</Badge>
                      )}
                    </Group>

                    {exec.cmd && (
                      <CodeHighlighter theme={settings.theme} copyButton={false}>
                        {exec.cmd}
                      </CodeHighlighter>
                    )}

                    {exec.args && exec.args.length > 0 && (
                      <Stack gap={2}>
                        <Text size="sm" c="dimmed">Arguments</Text>
                        <Stack gap={2} ml="sm">
                          {exec.args.map((arg, argIndex) => (
                            <Text key={argIndex} size="sm" c="dimmed">- {arg}</Text>
                          ))}
                        </Stack>
                      </Stack>
                    )}
                  </Stack>
                </Timeline.Item>
              ))}
            </Timeline>
          </>
        ) : (
          <Text c="dimmed" size="sm">No executables defined.</Text>
        )}
      </Stack>
    </Card>
  );
}
