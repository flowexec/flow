// CodeHighlighter Configuration
// This file allows you to configure the CodeHighlighter component in one place

import { ThemeName } from "../../theme/types";

export interface CodeHighlighterConfig {
  defaultCopyButton: boolean;

  defaultTheme: string;

  styling: {
    padding: string;
    fontSize: string;
    lineHeight: string;
    backgroundColor: string;
    copyButtonStyle: {
      background: string;
      border: string;
      borderRadius: string;
      padding: string;
      fontSize: string;
    };
  };
}

export const themeMapper: Record<ThemeName, string> = {
  everforest: "default",
  dark: "dark",
  dracula: "dark",
  light: "default",
  "tokyo-night": "dark",
};

// Default configuration
export const defaultConfig: CodeHighlighterConfig = {
  defaultCopyButton: true,
  defaultTheme: "default",

  styling: {
    padding: "var(--mantine-spacing-md)",
    fontSize: "var(--mantine-font-size-sm)",
    lineHeight: "1.5",
    backgroundColor: "var(--mantine-color-gray-7)",
    copyButtonStyle: {
      background: "rgba(255, 255, 255, 0.1)",
      border: "1px solid rgba(255, 255, 255, 0.2)",
      borderRadius: "var(--mantine-radius-xs)",
      padding: "2px 6px",
      fontSize: "10px",
    },
  },
};

// Dark theme configuration
export const darkThemeConfig: CodeHighlighterConfig = {
  ...defaultConfig,
  defaultTheme: "dark",
  styling: {
    ...defaultConfig.styling,
    backgroundColor: "var(--mantine-color-appshell-0)",
    copyButtonStyle: {
      ...defaultConfig.styling.copyButtonStyle,
      background: "rgba(255, 255, 255, 0.15)",
      border: "1px solid rgba(255, 255, 255, 0.3)",
    },
  },
};

// Light theme configuration
export const lightThemeConfig: CodeHighlighterConfig = {
  ...defaultConfig,
  defaultTheme: "default",
  styling: {
    ...defaultConfig.styling,
    backgroundColor: "var(--mantine-color-gray-0)",
    copyButtonStyle: {
      ...defaultConfig.styling.copyButtonStyle,
      background: "rgba(0, 0, 0, 0.1)",
      border: "1px solid rgba(0, 0, 0, 0.2)",
    },
  },
};

// Function to get configuration based on theme
export function getConfigForTheme(theme?: ThemeName): CodeHighlighterConfig {
  if (!theme) return defaultConfig;

  const prismTheme = themeMapper[theme] || "default";

  switch (prismTheme) {
    case "dark":
      return darkThemeConfig;
    case "default":
    default:
      return lightThemeConfig;
  }
}

export const currentConfig = defaultConfig;
