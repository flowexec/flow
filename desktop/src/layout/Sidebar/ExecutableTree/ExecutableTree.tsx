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
import React from "react";
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
import {useAppContext} from "../../../hooks/useAppContext.tsx";
import {Link} from "react-router";

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
  let icon: React.ReactNode;
  if (customNode.isNamespace && hasChildren) {
    if (selected && expanded) {
      icon = <IconFolderOpen size={16} />;
    } else {
      icon = <IconFolder size={16} />;
    }
  } else {
    switch (customNode.verbType) {
      case DeactivationVerbType:
        icon = <IconOctagon size={16} />;
        break;
      case ConfigurationVerbType:
        icon = <IconSettingsAutomation size={16} />;
        break;
      case DestructionVerbType:
        icon = <IconProgressX size={16} />;
        break;
      case RetrievalVerbType:
        icon = <IconProgressDown size={16} />;
        break;
      case UpdateVerbType:
        icon = <IconRefresh size={16} />;
        break;
      case ValidationVerbType:
        icon = <IconCircleCheckFilled size={16} />;
        break;
      case LaunchVerbType:
        icon = <IconWindowMaximize size={16} />;
        break;
      case CreationVerbType:
        icon = <IconCirclePlus size={16} />;
        break;
      case RestartVerbType:
        icon = <IconReload size={16} />;
        break;
      case BuildVerbType:
        icon = <IconBlocks size={16} />;
        break;
      default:
        icon = <IconPlayerPlayFilled size={16} />;
    }
  }

  if (customNode.isNamespace) {
    return (
      <Group gap="xs" {...elementProps} key={customNode.value} mb="3">
        {icon}
        <Text>{customNode.label}</Text>
      </Group>
    );
  }

  return (
    <Link to={`/executable/${customNode.value}`}>
      <Group gap="xs" {...elementProps} key={customNode.value} mb="3">
        {icon}
        <Text>{customNode.label}</Text>
      </Group>
    </Link>
  );
}

export function ExecutableTree() {
  const { executables } = useAppContext();
  const tree = useTree();

  // React.useEffect(() => {
  //   const selectedValue = tree.selectedState[0];
  //   if (selectedValue) {
  //     const findNode = (
  //       nodes: CustomTreeNodeData[]
  //     ): CustomTreeNodeData | undefined => {
  //       for (const node of nodes) {
  //         if (node.value === selectedValue) {
  //           return node;
  //         }
  //         if (node.children) {
  //           const found = findNode(node.children as CustomTreeNodeData[]);
  //           if (found) return found;
  //         }
  //       }
  //       return undefined;
  //     };
  //
  //     // const node = findNode(getTreeData(executables));
  //     // if (node && !node.isNamespace) {
  //     //   onSelectExecutable(selectedValue);
  //     // }
  //   }
  // }, [tree.selectedState, executables, onSelectExecutable]);

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
          <Tree
            data={getTreeData(executables)}
            selectOnClick
            tree={tree}
            renderNode={Leaf}
          />
        </ScrollArea>
      )}
    </>
  );
}
