import { MantineProvider, createTheme } from "@mantine/core";
import { useEffect } from "react";
import { useSettings } from "../hooks/useSettings";
import { themes } from "./themes";

interface ThemeProviderProps {
  children: React.ReactNode;
}

const createColorArray = (color: string) =>
  [
    color,
    color,
    color,
    color,
    color,
    color,
    color,
    color,
    color,
    color,
  ] as const;

export function ThemeProvider({ children }: ThemeProviderProps) {
  const { settings } = useSettings();
  const currentTheme = themes[settings.theme];

  // Set data-palette attribute for CSS variables
  useEffect(() => {
    document.documentElement.setAttribute("data-palette", settings.theme);
  }, [settings.theme]);

  const theme = createTheme({
    colors: {
      primary: createColorArray(currentTheme.colors.primary),
      secondary: createColorArray(currentTheme.colors.secondary),
      tertiary: createColorArray(currentTheme.colors.tertiary),
      info: createColorArray(currentTheme.colors.info),
      body: createColorArray(currentTheme.colors.body),
      border: createColorArray(currentTheme.colors.border),
      emphasis: createColorArray(currentTheme.colors.emphasis),
    },
    primaryColor: "primary",
    primaryShade: 0,
    defaultRadius: "md",
    fontFamily:
      'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    components: {
      Button: {
        defaultProps: {
          radius: "md",
        },
      },
      Card: {
        defaultProps: {
          radius: "md",
        },
      },
      ActionIcon: {
        defaultProps: {
          variant: "light",
          color: "primary",
        },
      },
      AppShell: {
        defaultProps: {
          styles: {
            main: {
              backgroundColor: currentTheme.colors.background.main,
              color: currentTheme.colors.body,
            },
            header: {
              backgroundColor: currentTheme.colors.background.header,
              color: currentTheme.colors.bodyLight || currentTheme.colors.body,
            },
            navbar: {
              backgroundColor: currentTheme.colors.background.sidebar,
              color: currentTheme.colors.bodyLight || currentTheme.colors.body,
            },
          },
        },
      },
    },
    other: {
      backgroundColors: {
        main: currentTheme.colors.background.main,
        sidebar: currentTheme.colors.background.sidebar,
        header: currentTheme.colors.background.header,
        card: currentTheme.colors.background.card,
      },
      colors: {
        body: currentTheme.colors.body,
        bodyLight: currentTheme.colors.bodyLight,
        border: currentTheme.colors.border,
      },
    },
  });

  return (
    <MantineProvider theme={theme} defaultColorScheme="dark">
      {children}
    </MantineProvider>
  );
}
