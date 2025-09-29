import { Anchor, Badge, Card, Divider, Group, SimpleGrid, Stack, Text, Title } from "@mantine/core";
import { IconArrowsRight } from "@tabler/icons-react";
import { useSettings } from "../../../hooks/useSettings";
import { EnrichedExecutable } from "../../../types/executable";
import { CodeHighlighter } from "../../../components/CodeHighlighter";
import { Link } from "wouter";

export type ExecutableParallelDetailsProps = {
  executable: EnrichedExecutable;
};

export function ExecutableParallelDetails({
  executable,
}: ExecutableParallelDetailsProps) {
  const { settings } = useSettings();

  const execs = executable.parallel?.execs || [];

  return (
    <Card withBorder shadow="sm" padding="md">
      <Stack gap="sm">
        <Group justify="space-between">
          <Title order={4}>
            <Group gap="xs">
              <IconArrowsRight size={16} />
              Parallel Configuration
            </Group>
          </Title>
          <Group gap="xs">
            {executable.parallel?.failFast && (
              <Badge variant="light" color={executable.parallel?.failFast ? "red.5" : "green.5"}>
                Fail fast: {executable.parallel?.failFast ? "on" : "off"}
              </Badge>
            )}
            {executable.parallel?.maxThreads && executable.parallel.maxThreads > 0 && (
              <Badge variant="light" color="blue.5">Max threads: {executable.parallel.maxThreads}</Badge>
            )}
            {execs.length > 0 && (
              <Badge variant="light" color="gray.5">Execs: {execs.length}</Badge>
            )}
          </Group>
        </Group>

        {execs.length > 0 ? (
          <>
            <Divider label="Executables" labelPosition="left" my="xs" />
            <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="md">
              {execs.map((exec, index) => (
                <Card key={index} withBorder padding="sm" radius="md">
                  <Stack gap={6}>
                    <Group justify="space-between" align="center">
                      <Group gap={6}>
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
                      <Group gap={6}>
                        {exec.retries !== undefined && exec.retries > 0 && (
                          <Badge size="sm" variant="dot" color="orange">retries: {exec.retries}</Badge>
                        )}
                        {exec.args && exec.args.length > 0 && (
                          <Badge size="sm" variant="light" color="gray">args: {exec.args.length}</Badge>
                        )}
                      </Group>
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
                </Card>
              ))}
            </SimpleGrid>
          </>
        ) : (
          <Text c="dimmed" size="sm">No executables defined.</Text>
        )}
      </Stack>
    </Card>
  );
}
