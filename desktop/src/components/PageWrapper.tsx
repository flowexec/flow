import { ScrollArea } from "@mantine/core";

export function PageWrapper({ children }: { children: React.ReactNode }) {
  return (
    <ScrollArea
      h="calc(100vh - var(--app-shell-padding-total))"
      w="100%"
      type="auto"
      scrollbarSize={6}
      scrollHideDelay={100}
      offsetScrollbars="present"
    >
      {children}
    </ScrollArea>
  );
}
