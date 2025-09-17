import { Box, Code } from "@mantine/core";
import Prism from "prismjs";
import "prismjs/components/prism-bash";
import "prismjs/components/prism-shell-session";
import { useEffect, useRef } from "react";
import { useNotifier } from "../../hooks/useNotifier";
import { ThemeName } from "../../theme/types";
import { NotificationType } from "../../types/notification";
import { getConfigForTheme, themeMapper } from "./config";

interface CodeHighlighterProps {
  children: string;
  className?: string;
  copyButton?: boolean;
  theme?: ThemeName;
}

export function CodeHighlighter({
  children,
  className,
  copyButton,
  theme,
}: CodeHighlighterProps) {
  const codeRef = useRef<HTMLElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const { setNotification } = useNotifier();
  const config = getConfigForTheme(theme);
  const finalCopyButton =
    copyButton !== undefined ? copyButton : config.defaultCopyButton;
  const language = "bash";

  // Dynamically load Prism theme CSS based on app theme via themeMapper
  useEffect(() => {
    const loadTheme = async () => {
      const variant = theme ? themeMapper[theme] ?? "default" : "default";
      if (variant === "dark") {
        await import("prismjs/themes/prism-dark.css");
      } else {
        await import("prismjs/themes/prism.css");
      }
    };
    void loadTheme();
  }, [theme]);

  useEffect(() => {
    if (codeRef.current) {
      Prism.highlightElement(codeRef.current);
    }
  }, [children]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(children);
      setNotification({
        title: "Code copied to clipboard",
        message: "The code has been copied to your clipboard.",
        type: NotificationType.Success,
        autoClose: true,
      });
    } catch (error) {
      setNotification({
        title: "Failed to copy code",
        message: "The code has not been copied to your clipboard.",
        type: NotificationType.Error,
        autoClose: true,
      });
      console.error(error);
    }
  };

  return (
    <Box className={className}>
      <Box
        ref={containerRef}
        pos="relative"
        style={{
          overflow: "hidden",
        }}
      >
        {finalCopyButton && (
          <Box pos="absolute" top={8} right={8} style={{ zIndex: 10 }}>
            <Code
              component="button"
              onClick={handleCopy}
              style={{
                cursor: "pointer",
                ...config.styling.copyButtonStyle,
                color: "inherit",
                fontFamily: "inherit",
              }}
            >
              Copy
            </Code>
          </Box>
        )}
        <pre
          style={{
            margin: 0,
            padding: config.styling.padding,
            background: config.styling.backgroundColor,
            border: "none",
            overflow: "auto",
            fontSize: config.styling.fontSize,
            lineHeight: config.styling.lineHeight,
          }}
        >
          <code
            ref={codeRef}
            className={`language-${language}`}
            style={{
              fontFamily: "var(--mantine-font-family-monospace)",
            }}
          >
            {children}
          </code>
        </pre>
      </Box>
    </Box>
  );
}
