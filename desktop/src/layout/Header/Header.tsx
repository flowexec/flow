import { ActionIcon } from "@mantine/core";
import { IconPlus, IconRefresh } from "@tabler/icons-react";
import styles from "./Header.module.css";
import {useAppContext} from "../../hooks/useAppContext.tsx";
import {NotificationType} from "../../types/notification.ts";
import {useNotifier} from "../../hooks/useNotifier.tsx";

export function Header() {
  const { refreshAll } = useAppContext();
  const { setNotification } = useNotifier();
  
  return (
    <div className={styles.header}>
      <div className={styles.header__actions}>
        <ActionIcon
          className={styles.header__button}
          onClick={() => {}}
          title="Create workspace"
          variant="light"
        >
          <IconPlus size={16} />
        </ActionIcon>

        <ActionIcon
          className={styles.header__button}
          variant="light"
          onClick={() => {
            refreshAll();
            setNotification({
              title: "Refresh completed",
              message: "flow data has synced and refreshed successfully",
              type: NotificationType.Success,
              autoClose: true,
              autoCloseDelay: 3000,
            });
          }}
          title="Refresh workspaces"
        >
          <IconRefresh size={16} />
        </ActionIcon>
      </div>
    </div>
  );
}
