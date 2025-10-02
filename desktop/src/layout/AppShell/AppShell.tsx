import {
  Loader,
  AppShell as MantineAppShell,
  Notification as MantineNotification,
  Text,
} from "@mantine/core";
import { ReactNode, useEffect, useState } from "react";
import { useAppContext } from "../../hooks/useAppContext.tsx";
import { useNotifier } from "../../hooks/useNotifier.tsx";
import { colorFromType, NotificationType } from "../../types/notification";
import { Sidebar } from "../Sidebar/Sidebar";
import styles from "./AppShell.module.css";

interface AppShellProps {
  children: ReactNode;
}

export function AppShell({ children }: AppShellProps) {
  const { isLoading, hasError } = useAppContext();
  const { notification, setNotification } = useNotifier();
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    if (hasError) {
      setNotification({
        title: "Unexpected error",
        message: hasError.message || "An error occurred",
        type: NotificationType.Error,
        autoClose: false,
      });
    }
  }, [hasError, setNotification]);

  return (
    <MantineAppShell
      header={{ height: "var(--app-header-height)" }}
      navbar={{ width: collapsed ? 56 : 250, breakpoint: "sm" }}
      padding="md"
      classNames={{
        root: styles.appShell,
        main: styles.main,
        header: styles.header,
        navbar: styles.navbar,
      }}
    >
      <MantineAppShell.Navbar>
        <Sidebar collapsed={collapsed} onToggle={() => setCollapsed((c) => !c)} />
      </MantineAppShell.Navbar>

      <MantineAppShell.Main>
        {hasError ? (
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              height: "100%",
              flexDirection: "column",
              gap: "1rem",
            }}
          >
            <Text c="red">Error loading data</Text>
            <Text c="red">{hasError.message}</Text>
          </div>
        ) : (
          <div style={{ position: "relative", height: "100%", flex: 1, minHeight: 0 }}>
            {children}
            {isLoading && (
              <div
                style={{
                  position: "absolute",
                  top: 16,
                  right: 16,
                  zIndex: 1000,
                }}
              >
                <Loader size="sm" />
              </div>
            )}
          </div>
        )}
      </MantineAppShell.Main>

      {notification && (
        <MantineNotification
          title={notification.title}
          onClose={() => setNotification(null)}
          color={colorFromType(notification.type)}
          style={{
            position: "fixed",
            bottom: 20,
            right: 20,
            zIndex: 1000,
          }}
        >
          {notification.message}
        </MantineNotification>
      )}
    </MantineAppShell>
  );
}
