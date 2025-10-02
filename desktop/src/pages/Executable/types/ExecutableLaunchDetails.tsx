import { Badge, Card, Code, Divider, Group, Stack, Text, Title } from "@mantine/core";
import { IconExternalLink } from "@tabler/icons-react";
import { EnrichedExecutable } from "../../../types/executable";

export type ExecutableLaunchDetailsProps = {
  executable: EnrichedExecutable;
};

export function ExecutableLaunchDetails({
  executable,
}: ExecutableLaunchDetailsProps) {
  return (
    <Card withBorder shadow="sm" padding="md">
      <Stack gap="sm">
        <Group justify="space-between" align="center">
          <Title order={4}>
            <Group gap="xs">
              <IconExternalLink size={16} />
              Launch Configuration
            </Group>
          </Title>
          <Group gap="xs">
            {typeof executable.launch?.app !== "undefined" && (
              <Badge variant="light" color={executable.launch?.app ? "blue.5" : "gray.5"}>
                app: {executable.launch?.app ? "set" : "unset"}
              </Badge>
            )}
          </Group>
        </Group>

        {executable.launch?.uri && (
          <>
            <Divider label="Target" labelPosition="left" my="xs" />
            <Text
              component="a"
              href={executable.launch.uri}
              target="_blank"
              rel="noopener noreferrer"
            >
              {executable.launch.uri}
            </Text>
          </>
        )}

        {executable.launch?.app && (
          <>
            <Divider label="Application" labelPosition="left" my="xs" />
            <Code>{executable.launch.app}</Code>
          </>
        )}
      </Stack>
    </Card>
  );
}
