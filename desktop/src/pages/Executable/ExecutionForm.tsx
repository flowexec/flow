import {
  Badge,
  Button,
  Group,
  Modal,
  Stack,
  Text,
  TextInput,
  Title,
} from "@mantine/core";
import { useState } from "react";
import { EnrichedExecutable } from "../../types/executable";
import { ExecutableArgument } from "../../types/generated/flowfile";

export interface ExecutionFormData {
  params: Record<string, string>;
  args: string;
}

export interface ExecutionFormProps {
  opened: boolean;
  onClose: () => void;
  onSubmit: (data: ExecutionFormData) => void;
  executable: EnrichedExecutable;
}

export function ExecutionForm({
  opened,
  onClose,
  onSubmit,
  executable,
}: ExecutionFormProps) {
  const [formData, setFormData] = useState<ExecutionFormData>({
    params: {},
    args: "",
  });

  // Get parameters that need prompts
  const promptParams =
    executable.exec?.params?.filter((param) => param.prompt) || [];

  // Get arguments
  const hasArgs = executable.exec?.args && executable.exec.args.length > 0;
  const args = executable.exec?.args || [];

  const handleSubmit = () => {
    onSubmit(formData);
    onClose();
  };

  const handleClose = () => {
    setFormData({ params: {}, args: "" });
    onClose();
  };

  const handleParamChange = (envKey: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      params: {
        ...prev.params,
        [envKey]: value,
      },
    }));
  };

  const handleArgsChange = (value: string) => {
    setFormData((prev) => ({
      ...prev,
      args: value,
    }));
  };

  const formatArgumentHint = (arg: ExecutableArgument): string => {
    const parts = [];

    if (arg.pos) {
      parts.push(`position ${arg.pos}`);
    } else if (arg.flag) {
      parts.push(`--${arg.flag}`);
    }

    if (arg.type && arg.type !== "string") {
      parts.push(`(${arg.type})`);
    }

    if (arg.required) {
      parts.push("required");
    } else if (arg.default) {
      parts.push(`default: ${arg.default}`);
    }

    return parts.join(" ");
  };

  const generateExampleArgs = (): string => {
    return args
      .map((arg) => {
        if (arg.pos) {
          return arg.default || `value${arg.pos}`;
        } else if (arg.flag) {
          if (arg.type === "bool") {
            return `${arg.flag}=${arg.default || "false"}`;
          } else {
            return `${arg.flag}=${arg.default || "value"}`;
          }
        }
        return "";
      })
      .filter(Boolean)
      .join(" ");
  };

  return (
    <Modal
      opened={opened}
      onClose={handleClose}
      title={`Execute ${executable.ref}`}
      size="lg"
    >
      <Stack gap="md">
        {promptParams.length > 0 && (
          <div>
            <Title order={4} mb="sm">
              Parameters
            </Title>
            <Stack gap="sm">
              {promptParams.map((param, index) => (
                <TextInput
                  key={index}
                  label={param.prompt}
                  value={formData.params[param.envKey || ""] || ""}
                  onChange={(event) => {
                    const target = event.target as HTMLInputElement;
                    if (target) {
                      handleParamChange(param.envKey || "", target.value);
                    }
                  }}
                  required={true}
                />
              ))}
            </Stack>
          </div>
        )}

        {hasArgs && (
          <div>
            <Title order={4} mb="sm">
              Arguments
            </Title>

            <Stack gap="xs" mb="md">
              {args.map((arg, index) => (
                <Group key={index} gap="xs" wrap="nowrap">
                  <Badge
                    size="sm"
                    variant="light"
                    color={arg.required ? "red" : "blue"}
                  >
                    {arg.envKey}
                  </Badge>
                  <Text size="xs" c="dimmed" style={{ flex: 1 }}>
                    {formatArgumentHint(arg)}
                  </Text>
                </Group>
              ))}
            </Stack>

            <Text size="sm" c="dimmed" mb="sm" mt="md">
              Enter all arguments as space-separated values:
            </Text>

            <TextInput
              placeholder={generateExampleArgs()}
              value={formData.args}
              onChange={(event) => {
                const target = event.target as HTMLInputElement;
                if (target) {
                  handleArgsChange(target.value);
                }
              }}
            />
          </div>
        )}

        <Group justify="flex-end" mt="md">
          <Button variant="light" onClick={handleClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit}>Execute</Button>
        </Group>
      </Stack>
    </Modal>
  );
}
