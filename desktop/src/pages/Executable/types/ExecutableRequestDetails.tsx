import { Badge, Card, Code, Group, Stack, Text, Title } from "@mantine/core";
import { IconExternalLink } from "@tabler/icons-react";
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
    <Card withBorder>
      <Stack gap="sm">
        <Title order={4}>
          <Group gap="xs">
            <IconExternalLink size={16} />
            Request Configuration
          </Group>
        </Title>
        <Stack gap="xs">
          {executable.request?.method && (
            <div>
              <Title order={5}>Method:</Title>
              <Badge variant="light">{executable.request.method}</Badge>
            </div>
          )}
          {executable.request?.url && (
            <div>
              <Title order={5}>URL:</Title>
              <Text
                component="a"
                href={executable.request.url}
                target="_blank"
                rel="noopener noreferrer"
              >
                {executable.request.url}
              </Text>
            </div>
          )}
          {executable.request?.timeout && (
            <div>
              <Title order={5}>Request Timeout:</Title>
              <Code>{executable.request.timeout}</Code>
            </div>
          )}
          {executable.request?.logResponse && (
            <div>
              <Title order={5}>Log Response:</Title>
              <Badge variant="light" color="blue">
                enabled
              </Badge>
            </div>
          )}
          {executable.request?.body && (
            <div>
              <Title order={5}>Body:</Title>
              <CodeHighlighter theme={settings.theme} copyButton={false}>
                {executable.request.body}
              </CodeHighlighter>
            </div>
          )}
          {executable.request?.headers &&
            Object.keys(executable.request.headers).length > 0 && (
              <div>
                <Title order={5}>Headers:</Title>
                <Stack gap="xs">
                  {Object.entries(executable.request.headers).map(
                    ([key, value]) => (
                      <div key={key}>
                        <Text size="sm" fw={500}>
                          {key}:
                        </Text>
                        <Code>{value}</Code>
                      </div>
                    ),
                  )}
                </Stack>
              </div>
            )}
          {executable.request?.validStatusCodes &&
            executable.request.validStatusCodes.length > 0 && (
              <div>
                <Title order={5}>Accepted Status Codes:</Title>
                <Group gap="xs">
                  {executable.request.validStatusCodes.map((code) => (
                    <Badge key={code} variant="light">
                      {code}
                    </Badge>
                  ))}
                </Group>
              </div>
            )}
          {executable.request?.responseFile && (
            <div>
              <Title order={5}>Response Saved To:</Title>
              <Code>{executable.request.responseFile.filename}</Code>
              {executable.request.responseFile.saveAs && (
                <div>
                  <Title order={6}>Response Saved As:</Title>
                  <Code>{executable.request.responseFile.saveAs}</Code>
                </div>
              )}
            </div>
          )}
          {executable.request?.transformResponse && (
            <div>
              <Title order={5}>Transformation Expression:</Title>
              <CodeHighlighter theme={settings.theme} copyButton={false}>
                {executable.request.transformResponse}
              </CodeHighlighter>
            </div>
          )}
        </Stack>
      </Stack>
    </Card>
  );
}
