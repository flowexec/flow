import React from 'react';
import { Box, Container, Group, Stack } from '@mantine/core';
import classes from './Hero.module.css';

interface BasePatternProps extends React.ComponentPropsWithoutRef<'svg'> {
  size?: number;
  className?: string;
}

type HeroVariant = 'center' | 'split' | 'left';
type PatternType = 'lines' | 'subtle' | 'none';

interface HeroProps {
  children: React.ReactNode;
  variant?: HeroVariant;
  pattern?: PatternType | React.ComponentType<BasePatternProps>;
  patternProps?: Partial<BasePatternProps>;
  className?: string;
  containerSize?: string | number;
}

interface HeroHeaderProps {
  children: React.ReactNode;
  className?: string;
}

interface HeroContentProps {
  children: React.ReactNode;
  className?: string;
}

interface HeroActionsProps {
  children: React.ReactNode;
  className?: string;
  align?: 'left' | 'center' | 'right';
}

// Compound component parts
function HeroHeader({ children, className }: HeroHeaderProps) {
  return (
    <Box className={`${classes.header} ${className ?? ''}`}>
      {children}
    </Box>
  );
}

function HeroContent({ children, className }: HeroContentProps) {
  return (
    <Box className={`${classes.content} ${className ?? ''}`}>
      {children}
    </Box>
  );
}

function HeroActions({ children, className, align = 'right' }: HeroActionsProps) {
  return (
    <Group className={`${classes.actions} ${classes[`actions-${align}`]} ${className ?? ''}`} gap="sm">
      {children}
    </Group>
  );
}

// Main Hero component
export function Hero({
  children,
  variant = 'center',
  pattern = 'subtle',
  patternProps,
  className,
  containerSize = 'xl',
}: HeroProps) {
  const renderPattern = () => {
    if (pattern === 'none') return null;

    if (typeof pattern === 'string') {
      // Use built-in subtle pattern
      return (
        <Box className={classes.pattern} aria-hidden="true">
          <Box className={`${classes.patternSubtle} ${classes[`pattern-${pattern}`]}`} />
        </Box>
      );
    }

    // Use custom pattern component
    const Pattern = pattern;
    return (
      <Box className={classes.pattern} aria-hidden="true">
        <Pattern size={600} className={classes.patternSvg} {...patternProps} />
      </Box>
    );
  };

  const getLayoutClasses = () => {
    const baseClasses = [classes.root, classes[`variant-${variant}`]];
    if (className) baseClasses.push(className);
    return baseClasses.join(' ');
  };

  const renderContent = () => {
    if (variant === 'split') {
      const headerChild = React.Children.toArray(children).find(
        child => React.isValidElement(child) && child.type === HeroHeader
      );
      const actionsChild = React.Children.toArray(children).find(
        child => React.isValidElement(child) && child.type === HeroActions
      );
      const otherChildren = React.Children.toArray(children).filter(
        child => React.isValidElement(child) &&
        child.type !== HeroHeader &&
        child.type !== HeroActions
      );

      return (
        <Group className={classes.splitLayout} align="flex-start" justify="space-between">
          <Stack gap="xs" className={classes.splitContent}>
            {headerChild}
            {otherChildren}
          </Stack>
          <Box className={classes.splitActions}>
            {actionsChild}
          </Box>
        </Group>
      );
    }

    return (
      <Stack gap="md" className={classes.stackedContent} align={variant === 'center' ? 'center' : 'flex-start'}>
        {children}
      </Stack>
    );
  };

  return (
    <Box className={getLayoutClasses()} component="section">
      {renderPattern()}

      <Container size={containerSize} className={classes.contentContainer}>
        <Box className={classes.inner}>
          {renderContent()}
        </Box>
      </Container>
    </Box>
  );
}

// Attach compound components
Hero.Header = HeroHeader;
Hero.Content = HeroContent;
Hero.Actions = HeroActions;

// Export types for external use
export type { HeroProps, HeroHeaderProps, HeroContentProps, HeroActionsProps, HeroVariant, PatternType };
