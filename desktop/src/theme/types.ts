export interface ColorPalette {
  primary: string;
  secondary: string;
  tertiary: string;
  info: string;
  body: string;
  bodyLight?: string;
  border: string;
  emphasis: string;
}

export interface Theme {
  name: ThemeName;
  darkMode: boolean;
  colors: {
    primary: string;
    secondary: string;
    tertiary: string;
    info: string;
    body: string;
    bodyLight?: string;
    border: string;
    emphasis: string;
    background: {
      main: string;
      sidebar: string;
      header: string;
      card: string;
      hero: string;
    };
  };
}

export type ThemeName =
  | "everforest"
  | "dark"
  | "dracula"
  | "light"
  | "tokyo-night";
