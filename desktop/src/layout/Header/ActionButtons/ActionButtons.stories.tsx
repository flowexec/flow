import type { Meta, StoryObj } from "@storybook/react";
import { ActionButtons } from "./ActionButtons";

const meta = {
  title: "Components/Header/ActionButtons",
  component: ActionButtons,
  parameters: {
    layout: "centered",
  },
  tags: ["autodocs"],
} satisfies Meta<typeof ActionButtons>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    onCreateWorkspace: () => console.log("Create workspace clicked"),
    onRefreshWorkspaces: () => console.log("Refresh workspaces clicked"),
  },
};

export const Interactive: Story = {
  args: {
    onCreateWorkspace: () => alert("Create workspace action triggered"),
    onRefreshWorkspaces: () => alert("Refresh workspaces action triggered"),
  },
};
