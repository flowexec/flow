import {ScrollArea} from "@mantine/core";

export function PageWrapper({children}: { children: React.ReactNode }) {
    return (
        <ScrollArea
            h="calc(100vh - var(--app-header-height) - var(--app-shell-padding-total))"
            w="calc(100vw - var(--app-navbar-width) - var(--app-shell-padding-total))"
            type="auto"
            scrollbarSize={6}
            scrollHideDelay={100}
            offsetScrollbars
        >
            {children}
        </ScrollArea>
    );
}