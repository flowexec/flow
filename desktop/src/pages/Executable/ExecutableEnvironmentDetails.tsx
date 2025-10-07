import {
  Badge,
  Card,
  Code,
  Grid,
  Group,
  Stack,
  Table,
  Text,
  Title,
} from "@mantine/core";
import { IconFlag, IconKey, IconFile } from "@tabler/icons-react";
import { EnrichedExecutable } from "../../types/executable";
import {
  ExecutableArgument,
  ExecutableParameter,
} from "../../types/generated/flowfile";

export type ExecutableEnvironmentDetailsProps = {
  executable: EnrichedExecutable;
};

type ParamType = "static" | "secret" | "prompt" | "file" | "unknown";

export function ExecutableEnvironmentDetails({executable}: ExecutableEnvironmentDetailsProps) {
  const env =
    executable.exec ||
    executable.launch ||
    executable.request ||
    executable.render ||
    executable.serial ||
    executable.parallel;

  const hasParams = env?.params && env.params.length > 0;
  const hasArgs = env?.args && env.args.length > 0;

  if (!env || (!hasParams && !hasArgs)) return null;

  return (
    <Grid>
      {env.params && env.params.length > 0 && (
        <Grid.Col span={12}>
          <Card withBorder>
            <Stack gap="sm">
              <Title order={4}>
                <Group gap="xs">
                  <IconKey size={16} />
                  Environment Parameters
                </Group>
              </Title>
              <Table withTableBorder>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Destination</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th>Source</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {env.params.map((param: ExecutableParameter, index: number) => {
                    const type = getParamType(param);
                    const source = getParamSource(param);
                    const fileDestination = param.outputFile;

                    return (
                      <Table.Tr key={index}>
                        <Table.Td>
                          {param.envKey ? (
                            <Group gap={4} wrap="nowrap">
                              <Code>{param.envKey}</Code>
                              <Badge size="xs" variant="dot" color="secondary">ENV</Badge>
                            </Group>
                          ) : fileDestination ? (
                            <Group gap={4} wrap="nowrap">
                              <IconFile size={14} />
                              <Code>{fileDestination}</Code>
                            </Group>
                          ) : (
                            <Text c="dimmed" size="sm">-</Text>
                          )}
                        </Table.Td>
                        <Table.Td>
                          <Text
                            size="sm"
                            variant="transparent"
                            c={getParamTypeColor(type)}
                          >
                            {type}
                          </Text>
                        </Table.Td>
                        <Table.Td>
                          <Text size="sm" style={{ wordBreak: "break-word", maxWidth: 300 }} truncate>
                            {source}
                          </Text>
                        </Table.Td>
                      </Table.Tr>
                    );
                  })}
                </Table.Tbody>
              </Table>
            </Stack>
          </Card>
        </Grid.Col>
      )}

      {env.args && env.args.length > 0 && (
        <Grid.Col span={12}>
          <Card withBorder>
            <Stack gap="sm">
              <Title order={4}>
                <Group gap="xs">
                  <IconFlag size={16} />
                  Command Arguments
                </Group>
              </Title>
              <Table withTableBorder>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Destination</Table.Th>
                    <Table.Th>CLI Input</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th>Required</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {env.args.map((arg: ExecutableArgument, index: number) => {
                    const fileDestination = arg.outputFile;

                    return (
                      <Table.Tr key={index}>
                        <Table.Td>
                          {arg.envKey ? (
                            <Group gap={4} wrap="nowrap">
                              <Code>{arg.envKey}</Code>
                              <Badge size="xs" variant="dot" color="secondary">ENV</Badge>
                            </Group>
                          ) : fileDestination ? (
                            <Group gap={4} wrap="nowrap">
                              <IconFile size={14} />
                              <Code>{fileDestination}</Code>
                            </Group>
                          ) : (
                            <Text c="dimmed" size="sm">Not saved</Text>
                          )}
                        </Table.Td>
                        <Table.Td>
                          <Group gap="xs" wrap="wrap">
                            <Code>
                              {arg.pos ? `position=${arg.pos}` : `flag=${arg.flag}`}
                            </Code>
                            {arg.default && (
                              <Text size="xs" c="secondary">
                                default: {arg.default}
                              </Text>
                            )}
                          </Group>
                        </Table.Td>
                        <Table.Td>
                          <Code>{arg.type || "string"}</Code>
                        </Table.Td>
                        <Table.Td>
                          <Badge
                            size="sm"
                            variant="light"
                            color={arg.required ? "red.5" : "green.5"}
                          >
                            {arg.required ? "Yes" : "No"}
                          </Badge>
                        </Table.Td>
                      </Table.Tr>
                    );
                  })}
                </Table.Tbody>
              </Table>
            </Stack>
          </Card>
        </Grid.Col>
      )}
    </Grid>
  );
}

function getParamType(param: ExecutableParameter): ParamType {
  if (param.text) return "static";
  if (param.secretRef) return "secret";
  if (param.prompt) return "prompt";
  if (param.envFile) return "file";
  return "unknown";
}

function getParamSource(param: ExecutableParameter): string {
  return param.text || param.secretRef || param.prompt || param.envFile || "-";
}

function getParamTypeColor(type: ParamType): string {
  switch (type) {
    case "secret": return "red.5";
    case "prompt": return "blue.5";
    case "file": return "purple.5";
    case "static": return "gray.5";
    default: return "gray.5";
  }
}
