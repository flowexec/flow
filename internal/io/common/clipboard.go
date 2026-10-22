package common

import (
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
)

// CopySecretToClipboard copies text the way tuikit's Container.CopyToClipboard does (native
// clipboard and OSC 52) but confirms with successMsg instead of tuikit's toast, which previews
// the copied text and would put a secret value on screen.
func CopySecretToClipboard(container *tuikit.Container, text, successMsg string) {
	container.Send(tea.SetClipboard(text), 0)
	// In a goroutine since pbcopy/xclip can hang under raw terminal mode. A failure is expected
	// where there is no native clipboard (e.g. over SSH); the OSC 52 copy still applies there.
	go func() {
		_ = clipboard.WriteAll(text)
		container.SetNotice(successMsg, themes.OutputLevelSuccess)
	}()
}
