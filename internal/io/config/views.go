package config

import (
	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/v2/types/config"
)

func NewUserConfigView(
	container *tuikit.Container,
	cfg config.Config,
) tuikit.View {
	return views.NewEntityView(container.RenderState(), &cfg, types.EntityFormatDocument)
}
