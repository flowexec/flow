import {
  Group,
  RenderTreeNodePayload,
  ScrollArea,
  Text,
  Tree,
  TreeNodeData,
  useTree,
} from "@mantine/core";
import {
  IconBlocks,
  IconCircleCheckFilled,
  IconCirclePlus,
  IconFolder,
  IconFolderOpen,
  IconOctagon,
  IconPlayerPlayFilled,
  IconProgressDown,
  IconProgressX,
  IconRefresh,
  IconReload,
  IconSettingsAutomation,
  IconWindowMaximize,
} from "@tabler/icons-react";
import { useMemo, useCallback } from "react";
import {
  BuildVerbType,
  ConfigurationVerbType,
  CreationVerbType,
  DeactivationVerbType,
  DestructionVerbType,
  EnrichedExecutable,
  GetUIVerbType,
  LaunchVerbType,
  RestartVerbType,
  RetrievalVerbType,
  UpdateVerbType,
  ValidationVerbType,
} from "../../../types/executable";
import { useAppContext } from "../../../hooks/useAppContext.tsx";
import { useLocation } from "wouter";

interface CustomTreeNodeData extends TreeNodeData {
  isNamespace: boolean;
  verbType: string | null;
}

function getTreeData(executables: EnrichedExecutable[]): CustomTreeNodeData[] {
  const execsByNamespace: Record<string, EnrichedExecutable[]> = {};
  const rootExecutables: EnrichedExecutable[] = [];

  // Separate executables into namespaced and root level
  for (const executable of executables) {
    if (executable.namespace) {
      if (!execsByNamespace[executable.namespace]) {
        execsByNamespace[executable.namespace] = [];
      }
      execsByNamespace[executable.namespace].push(executable);
    } else {
      rootExecutables.push(executable);
    }
  }

  const treeData: CustomTreeNodeData[] = [];

  Object.entries(execsByNamespace)
    .sort(([namespaceA], [namespaceB]) => namespaceA.localeCompare(namespaceB))
    .forEach(([namespace, executables]) => {
      treeData.push({
        label: namespace,
        value: namespace,
        isNamespace: true,
        verbType: null,
        children: executables
          .sort((a, b) => (a.id || "").localeCompare(b.id || ""))
          .map((executable) => ({
            label: executable.name
              ? executable.verb + " " + executable.name
              : executable.verb,
            value: executable.ref,
            isNamespace: false,
            verbType: GetUIVerbType(executable),
          })),
      });
    });

  rootExecutables
    .sort((a, b) => (a.id || "").localeCompare(b.id || ""))
    .forEach((executable) => {
      treeData.push({
        label: executable.name
          ? executable.verb + " " + executable.name
          : executable.verb,
        value: executable.ref,
        isNamespace: false,
        verbType: GetUIVerbType(executable),
      });
    });

  return treeData;
}

function Leaf({
  node,
  selected,
  expanded,
  hasChildren,
  elementProps,
}: RenderTreeNodePayload) {
  const customNode = node as CustomTreeNodeData;
  const [, setLocation] = useLocation();

  const icon = useMemo(() => {
    if (customNode.isNamespace && hasChildren) {
      if (selected && expanded) {
        return <IconFolderOpen size={16} />;
      } else {
        return <IconFolder size={16} />;
      }
    } else {
      switch (customNode.verbType) {
        case DeactivationVerbType:
          return <IconOctagon size={16} />;
        case ConfigurationVerbType:
          return <IconSettingsAutomation size={16} />;
        case DestructionVerbType:
          return <IconProgressX size={16} />;
        case RetrievalVerbType:
          return <IconProgressDown size={16} />;
        case UpdateVerbType:
          return <IconRefresh size={16} />;
        case ValidationVerbType:
          return <IconCircleCheckFilled size={16} />;
        case LaunchVerbType:
          return <IconWindowMaximize size={16} />;
        case CreationVerbType:
          return <IconCirclePlus size={16} />;
        case RestartVerbType:
          return <IconReload size={16} />;
        case BuildVerbType:
          return <IconBlocks size={16} />;
        default:
          return <IconPlayerPlayFilled size={16} />;
      }
    }
  }, [hasChildren, selected, expanded]);

  const handleExecutableClick = useCallback(() => {
    const encodedId = encodeURIComponent(customNode.value);
    setLocation(`/executable/${encodedId}`);
  }, [setLocation]);

  if (customNode.isNamespace) {
    return (
      <Group gap="xs" {...elementProps} key={customNode.value} mb="3">
        {icon}
        <Text>{customNode.label}</Text>
      </Group>
    );
  }

  return (
    <Group
      gap="xs"
      {...elementProps}
      mb="3"
      onClick={handleExecutableClick}
      style={{ cursor: "pointer" }}
    >
      {icon}
      <Text>{customNode.label}</Text>
    </Group>
  );
}

export function ExecutableTree() {
  const { executables } = useAppContext();
  const tree = useTree();

  const treeData = useMemo(() => getTreeData(executables), [executables]);

  return (
    <>
      <Text size="xs" fw={700} c="dimmed" mb="0" mt="md">
        EXECUTABLES ({executables.length})
      </Text>
      {executables.length === 0 ? (
        <Text size="xs" c="red">
          No executables found
        </Text>
      ) : (
        <ScrollArea scrollbarSize={6} scrollHideDelay={100}>
          <Tree data={treeData} selectOnClick tree={tree} renderNode={Leaf} />
        </ScrollArea>
      )}
    </>
  );
}
