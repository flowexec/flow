package logger

import "github.com/flowexec/tuikit/themes"

func Theme(name string) themes.Theme {
	theme := themes.EverforestTheme()
	themeFunc, ok := themes.AllThemes()[name]
	if ok {
		theme = themeFunc()
	}
	return theme
}
