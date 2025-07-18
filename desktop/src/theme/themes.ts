import { Theme, ThemeName } from "./types";

export const themes: Record<ThemeName, Theme> = {
  everforest: {
    name: "everforest",
    darkMode: true,
    colors: {
      primary: "#7FBBB3",
      secondary: "#83C092",
      tertiary: "#A7C080",
      info: "#7FBBB3",
      body: "#343F44",
      bodyLight: "#E6E0D5",
      border: "#5C6A72",
      emphasis: "#DBBC7F",
      background: {
        main: "#F0F0F0",
        sidebar: "#1E2326",
        header: "#1E2326",
        card: "#F8F8F8",
      },
    },
  },
  dark: {
    name: "dark",
    darkMode: true,
    colors: {
      primary: "#39",
      secondary: "#228",
      tertiary: "#63",
      info: "#39",
      body: "#E6EDF3",
      bodyLight: "#C9D1D9",
      border: "#240",
      emphasis: "#30",
      background: {
        main: "#1A1B1E",
        sidebar: "#141517",
        header: "#141517",
        card: "#25262B",
      },
    },
  },
  dracula: {
    name: "dracula",
    darkMode: true,
    colors: {
      primary: "#bd93f9",
      secondary: "#8be9fd",
      tertiary: "#ffb86c",
      info: "#bd93f9",
      body: "#f8f8f2",
      bodyLight: "#E2E2DC",
      border: "#6272A4",
      emphasis: "#f1fa8c",
      background: {
        main: "#282A36",
        sidebar: "#1E1F29",
        header: "#1E1F29",
        card: "#44475A",
      },
    },
  },
  light: {
    name: "light",
    darkMode: false,
    colors: {
      primary: "#0969DA",
      secondary: "#1F883D",
      tertiary: "#8250DF",
      info: "#0969DA",
      body: "#24292F",
      bodyLight: "#F6F8FA",
      border: "#D0D7DE",
      emphasis: "#0969DA",
      background: {
        main: "#F0F0F0",
        sidebar: "#2D333B",
        header: "#2D333B",
        card: "#F8F8F8",
      },
    },
  },
  "tokyo-night": {
    name: "tokyo-night",
    darkMode: true,
    colors: {
      primary: "#bb9af7",
      secondary: "#7aa2f7",
      tertiary: "#2ac3de",
      info: "#bb9af7",
      body: "#a9b1d6",
      bodyLight: "#C0C8E0",
      border: "#565f89",
      emphasis: "#7aa2f7",
      background: {
        main: "#1A1B26",
        sidebar: "#16161E",
        header: "#16161E",
        card: "#24283B",
      },
    },
  },
};
