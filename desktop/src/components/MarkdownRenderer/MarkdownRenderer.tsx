import { Code, Text, Title } from "@mantine/core";
import React, { ComponentPropsWithoutRef } from "react";
import ReactMarkdown from "react-markdown";
import { CodeHighlighter } from "../CodeHighlighter";
import { useSettings } from "../../hooks/useSettings";
import styles from "./MarkdownRenderer.module.css";

interface MarkdownRendererProps {
  children: string;
  className?: string;
}

export function MarkdownRenderer({
  children,
  className,
}: MarkdownRendererProps) {
  const { settings } = useSettings();

  return (
    <div className={`${styles.container} ${className || ""}`}>
      <ReactMarkdown
        components={{
          h1: (props: ComponentPropsWithoutRef<typeof Title>) => (
            <Title order={3} {...props} />
          ),
          h2: (props: ComponentPropsWithoutRef<typeof Title>) => (
            <Title order={4} {...props} />
          ),
          h3: (props: ComponentPropsWithoutRef<typeof Title>) => (
            <Title order={5} {...props} />
          ),
          h4: (props: ComponentPropsWithoutRef<typeof Title>) => (
            <Title order={6} {...props} />
          ),
          h5: (props: ComponentPropsWithoutRef<typeof Text>) => (
            <Text size="sm" fw={700} {...props} />
          ),
          h6: (props: ComponentPropsWithoutRef<typeof Text>) => (
            <Text size="xs" fw={500} {...props} />
          ),
          p: (props: ComponentPropsWithoutRef<typeof Text>) => (
            <Text size="sm" {...props} />
          ),
          code: (props: ComponentPropsWithoutRef<typeof Code>) => (
            <Code {...props} />
          ),
          pre: (props: ComponentPropsWithoutRef<typeof Code>) => {
            const codeElement = props.children as React.ReactElement;
            const codeContent = codeElement?.props?.children || "";
            return (
              <CodeHighlighter theme={settings.theme}>
                {codeContent}
              </CodeHighlighter>
            );
          },
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}
