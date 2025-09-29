import { Badge, Card, Code, Divider, Group, Stack, Table, Text, Title } from "@mantine/core";
import { IconArrowForwardUpDouble } from "@tabler/icons-react";
import { useSettings } from "../../../hooks/useSettings";
import { EnrichedExecutable } from "../../../types/executable";
import { CodeHighlighter } from "../../../components/CodeHighlighter";

export type ExecutableRequestDetailsProps = {
  executable: EnrichedExecutable;
};

export function ExecutableRequestDetails({
  executable,
}: ExecutableRequestDetailsProps) {
  const { settings } = useSettings();

  return (
    <Card withBorder shadow="sm" padding="md">
      <Stack gap="sm">
        <Group justify="space-between" align="center">
          <Title order={4}>
            <Group gap="xs">
              <IconArrowForwardUpDouble size={16} />
              Request Configuration
            </Group>
          </Title>
          <Group gap="xs">
            {executable.request?.method && (
              <Badge variant="light" color={methodColor(executable.request.method)}>{executable.request.method}</Badge>
            )}
            {executable.request?.timeout && (
              <Badge variant="light" color="gray.5">timeout: {executable.request?.timeout}</Badge>
            )}
            <Badge variant="light" color={executable.request?.logResponse ? "blue.5" : "gray.5"}>
              log response: {executable.request?.logResponse ? "on" : "off"}
            </Badge>
          </Group>
        </Group>

        {executable.request?.url && (
          <>
            <Divider label="Target" labelPosition="left" my="xs" />
            <Text
              component="a"
              href={executable.request.url}
              target="_blank"
              rel="noopener noreferrer"
            >
              {executable.request.url}
            </Text>
          </>
        )}

        {executable.request?.headers && Object.keys(executable.request.headers).length > 0 && (
          <>
            <Divider label="Headers" labelPosition="left" my="xs" />
            <Table variant="vertical" withTableBorder layout="fixed">
              <Table.Tbody>
                {Object.entries(executable.request.headers).map(([key, value]) => (
                  <Table.Tr key={key}>
                    <Table.Th w="30%">{key}</Table.Th>
                    <Table.Td>
                      <Code>{String(value)}</Code>
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </>
        )}

        {executable.request?.validStatusCodes && executable.request.validStatusCodes.length > 0 && (
          <>
            <Divider label="Accepted status codes" labelPosition="left" my="xs" />
            <Group gap="xs">
              {executable.request.validStatusCodes.map((code) => (
                <Badge key={code} variant="light">{code}</Badge>
              ))}
            </Group>
          </>
        )}

        {executable.request?.body && (
          <>
            <Divider label="Body" labelPosition="left" my="xs" />
            <CodeHighlighter theme={settings.theme} copyButton={false}>
              {executable.request.body}
            </CodeHighlighter>
          </>
        )}

        {executable.request?.transformResponse && (
          <>
            <Divider label="Transform response" labelPosition="left" my="xs" />
            <CodeHighlighter theme={settings.theme} copyButton={false}>
              {executable.request.transformResponse}
            </CodeHighlighter>
          </>
        )}

        {executable.request?.responseFile && (
          <>
            <Divider label="Response file" labelPosition="left" my="xs" />
            <Group gap="md">
              <Stack gap={2}>
                <Text size="sm" c="dimmed">Filename</Text>
                <Code>{executable.request.responseFile.filename}</Code>
              </Stack>
              {executable.request.responseFile.saveAs && (
                <Stack gap={2}>
                  <Text size="sm" c="dimmed">Save as</Text>
                  <Code>{executable.request.responseFile.saveAs}</Code>
                </Stack>
              )}
            </Group>
          </>
        )}
      </Stack>
    </Card>
  );
}

const methodColor = (method?: string) => {
  switch ((method || "").toUpperCase()) {
    case "GET":
      return "green.5";
    case "POST":
      return "blue.5";
    case "PUT":
      return "orange.5";
    case "PATCH":
      return "yellow.5";
    case "DELETE":
      return "red.5";
    default:
      return "gray.5";
  }
};
