import {
  ActionIcon,
  Badge,
  Card,
  Code,
  Grid,
  Group,
  Stack,
  Table,
  Text,
  Title,
  Tooltip,
} from "@mantine/core";
import { IconFlag, IconInfoCircle, IconKey } from "@tabler/icons-react";
import { EnrichedExecutable } from "../../types/executable";
import {
  ExecutableArgument,
  ExecutableParameter,
} from "../../types/generated/flowfile";

export type ExecutableEnvironmentDetailsProps = {
  executable: EnrichedExecutable;
};

export function ExecutableEnvironmentDetails({
  executable,
}: ExecutableEnvironmentDetailsProps) {
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
              <Table>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Variable</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th>Source</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {env.params.map(
                    (param: ExecutableParameter, index: number) => {
                      const type = param.text
                        ? "static"
                        : param.secretRef
                          ? "secret"
                          : "prompt";
                      const source =
                        param.text || param.secretRef || param.prompt;

                      return (
                        <Table.Tr key={index}>
                          <Table.Td>
                            <Code>{param.envKey}</Code>
                          </Table.Td>
                          <Table.Td>
                            <Badge
                              size="sm"
                              variant="light"
                              color={
                                type === "secret"
                                  ? "red"
                                  : type === "prompt"
                                    ? "blue"
                                    : "gray"
                              }
                            >
                              {type}
                            </Badge>
                          </Table.Td>
                          <Table.Td>
                            <Text size="sm" style={{ maxWidth: 200 }}>
                              {source}
                            </Text>
                          </Table.Td>
                        </Table.Tr>
                      );
                    },
                  )}
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
              <Table>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Variable</Table.Th>
                    <Table.Th>Input</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th>Required</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {env.args.map((arg: ExecutableArgument, index: number) => (
                    <Table.Tr key={index}>
                      <Table.Td>
                        <Code>{arg.envKey}</Code>
                      </Table.Td>
                      <Table.Td>
                        <Group gap="xs">
                          <Badge size="sm" variant="light">
                            {arg.pos ? `pos-${arg.pos}` : `--${arg.flag}`}
                          </Badge>
                          {arg.default && (
                            <Tooltip label={`Default: ${arg.default}`}>
                              <ActionIcon variant="subtle" size="xs">
                                <IconInfoCircle size={12} />
                              </ActionIcon>
                            </Tooltip>
                          )}
                        </Group>
                      </Table.Td>
                      <Table.Td>
                        <Text size="sm">{arg.type || "string"}</Text>
                      </Table.Td>
                      <Table.Td>
                        <Badge
                          size="sm"
                          variant="light"
                          color={arg.required ? "red" : "green"}
                        >
                          {arg.required ? "Yes" : "No"}
                        </Badge>
                      </Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Tbody>
              </Table>
            </Stack>
          </Card>
        </Grid.Col>
      )}
    </Grid>
  );
}
