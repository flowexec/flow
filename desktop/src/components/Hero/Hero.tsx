import { Box, Container, Group, Stack } from "@mantine/core";
import React from "react";
import classes from "./Hero.module.css";

export type HeroVariant = "center" | "split" | "left";
export type PatternType = "lines" | "subtle" | "none";

export interface HeroProps {
  children: React.ReactNode;
  variant?: HeroVariant;
  pattern?: PatternType;
  className?: string;
  containerSize?: string | number;
}

export interface HeroHeaderProps {
  children: React.ReactNode;
  className?: string;
}

export interface HeroContentProps {
  children: React.ReactNode;
  className?: string;
}

export interface HeroActionsProps {
  children: React.ReactNode;
  className?: string;
  align?: "left" | "center" | "right";
}

export function HeroHeader({ children, className }: HeroHeaderProps) {
  return (
    <Box className={`${classes.header} ${className ?? ""}`}>{children}</Box>
  );
}

export function HeroContent({ children, className }: HeroContentProps) {
  return (
    <Box className={`${classes.content} ${className ?? ""}`}>{children}</Box>
  );
}

export function HeroActions({
  children,
  className,
  align = "right",
}: HeroActionsProps) {
  return (
    <Group
      className={`${classes.actions} ${classes[`actions-${align}`]} ${className ?? ""}`}
      gap="sm"
    >
      {children}
    </Group>
  );
}

export function Hero({
  children,
  variant = "center",
  pattern = "subtle",
  className,
  containerSize = "xl",
}: HeroProps) {
  const renderPattern = () => {
    if (pattern === "none") return null;
    return (
      <Box aria-hidden="true">
        <Box
          className={`${classes.pattern} ${classes[`pattern-${pattern}`]}`}
        />
      </Box>
    );
  };

  const getLayoutClasses = () => {
    const baseClasses = [classes.root, classes[`variant-${variant}`]];
    if (className) baseClasses.push(className);
    return baseClasses.join(" ");
  };

  const renderContent = () => {
    if (variant === "split") {
      const headerChild = React.Children.toArray(children).find(
        (child) => React.isValidElement(child) && child.type === HeroHeader
      );
      const actionsChild = React.Children.toArray(children).find(
        (child) => React.isValidElement(child) && child.type === HeroActions
      );
      const otherChildren = React.Children.toArray(children).filter(
        (child) =>
          React.isValidElement(child) &&
          child.type !== HeroHeader &&
          child.type !== HeroActions
      );

      return (
        <Group
          className={classes.splitLayout}
          align="flex-start"
          justify="space-between"
        >
          <Stack gap="xs" className={classes.splitContent}>
            {headerChild}
            {otherChildren}
          </Stack>
          <Box className={classes.splitActions}>{actionsChild}</Box>
        </Group>
      );
    }

    return (
      <Stack
        gap="md"
        className={classes.stackedContent}
        align={variant === "center" ? "center" : "flex-start"}
      >
        {children}
      </Stack>
    );
  };

  return (
    <Box className={getLayoutClasses()} component="section">
      {renderPattern()}

      <Container size={containerSize} className={classes.contentContainer}>
        <Box className={classes.inner}>{renderContent()}</Box>
      </Container>
    </Box>
  );
}

Hero.Header = HeroHeader;
Hero.Content = HeroContent;
Hero.Actions = HeroActions;
