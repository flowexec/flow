import { Box, Group, Paper, Text } from "@mantine/core";
import styles from "../pages/Settings/Settings.module.css";

interface SettingRowProps {
  label: string;
  description?: string;
  children: React.ReactNode;
}

export function SettingRow({ label, description, children }: SettingRowProps) {
  return (
    <Box py="md">
      <Group align="flex-start" gap="xl">
        <Box className={styles.settingLabel}>
          <Text size="sm" fw={500}>
            {label}
          </Text>
          {description && (
            <Text size="xs" c="dimmed" mt={2}>
              {description}
            </Text>
          )}
        </Box>
        <Box className={styles.settingControl}>{children}</Box>
      </Group>
    </Box>
  );
}

interface SettingSectionProps {
  title: string;
  children: React.ReactNode;
}

export function SettingSection({ title, children }: SettingSectionProps) {
  return (
    <Box mb="lg">
      <Text
        size="xs"
        fw={500}
        mb="xs"
        c="dimmed"
        tt="uppercase"
        className={styles.sectionTitle}
      >
        {title}
      </Text>
      <Paper className={styles.settingCard}>{children}</Paper>
    </Box>
  );
}
