import type { Meta, StoryObj } from "@storybook/react";
import { Header } from "./Header";
import { AppShell } from "@mantine/core";

const meta = {
  title: "Components/Header",
  component: Header,
  parameters: {
    layout: "fullscreen",
    docs: {
      story: {
        height: "100px",
      },
    },
  },
  decorators: [
    (Story) => (
      <AppShell header={{ height: 60 }}>
        <Story />
      </AppShell>
    ),
  ],
  tags: ["autodocs"],
} satisfies Meta<typeof Header>;

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
