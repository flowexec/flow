import "@mantine/core/styles.css";
import { ThemeProvider } from "../theme/ThemeProvider";
import { HeadScripts } from "../theme/HeadScripts";
import { SettingsProvider } from "../hooks/useSettings";

export function Layout({ children }: { children?: React.ReactNode }) {
  return (
    <SettingsProvider>
      <ThemeProvider>
        <HeadScripts />
        {children}
      </ThemeProvider>
    </SettingsProvider>
  );
}

