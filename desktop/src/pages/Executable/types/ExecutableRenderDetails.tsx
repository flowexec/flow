import { Card, Code, Divider, Group, Stack, Text, Title } from "@mantine/core";
import { IconArrowAutofitHeightFilled
} from "@tabler/icons-react";
import { EnrichedExecutable } from "../../../types/executable";

export type ExecutableRenderDetailsProps = {
  executable: EnrichedExecutable;
};

export function ExecutableRenderDetails({
  executable,
}: ExecutableRenderDetailsProps) {
  return (
    <Card withBorder shadow="sm" padding="md">
      <Stack gap="sm">
        <Group justify="space-between" align="center">
          <Title order={4}>
            <Group gap="xs">
              <IconArrowAutofitHeightFilled size={16} />
              Render Configuration
            </Group>
          </Title>
        </Group>

        {(executable.render?.templateFile || executable.render?.templateDataFile) && (
          <>
            <Divider label="Template" labelPosition="left" my="xs" />
            <Stack gap="xs">
              {executable.render?.templateFile && (
                <Stack gap={2}>
                  <Text size="xs" c="dimmed">Template file</Text>
                  <Code>{executable.render.templateFile}</Code>
                </Stack>
              )}
              {executable.render?.templateDataFile && (
                <Stack gap={2}>
                  <Text size="xs" c="dimmed">Template data file</Text>
                  <Code>{executable.render.templateDataFile}</Code>
                </Stack>
              )}
            </Stack>
          </>
        )}

        {executable.render?.dir && (
          <>
            <Divider label="Working directory" labelPosition="left" my="xs" />
            <Code>{executable.render.dir}</Code>
          </>
        )}
      </Stack>
    </Card>
  );
}
