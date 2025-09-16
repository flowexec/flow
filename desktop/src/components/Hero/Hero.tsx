import React from 'react';
import { Box, Center, Container } from '@mantine/core';
import classes from './Hero.module.css';

interface BasePatternProps extends React.ComponentPropsWithoutRef<'svg'> {
  size?: number;
  className?: string;
}

interface HeroProps {
  children: React.ReactNode;
  pattern: React.ComponentType<BasePatternProps>;
  patternProps?: Partial<BasePatternProps>;
  className?: string;
  containerSize?: string | number;
}

export function Hero({
  children,
  pattern: Pattern,
  patternProps,
  className,
  containerSize = 'xl',
}: HeroProps) {
  return (
    <Box className={`${classes.root} ${className ?? ''}`} component="section">
        <Box className={classes.pattern} aria-hidden="true">
          <Pattern size={600} className={classes.patternSvg} {...patternProps} />
        </Box>

      <Container size={containerSize} className={classes.contentContainer}>
        <Center className={classes.content}>
          <Box className={classes.inner}>{children}</Box>
        </Center>
      </Container>
    </Box>
  );
}
