package common

import (
	"github.com/atotto/clipboard"
	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
)

// CopyToClipboard writes text to the system clipboard in a goroutine
// to avoid blocking the bubbletea event loop (pbcopy/xclip can hang
// under raw terminal mode). Shows a toast notice on success or failure.
func CopyToClipboard(container *tuikit.Container, text, successMsg string) {
	go func() {
		if err := clipboard.WriteAll(text); err != nil {
			container.SetNotice("unable to copy to clipboard", themes.OutputLevelError)
		} else {
			container.SetNotice(successMsg, themes.OutputLevelSuccess)
		}
	}()
}
