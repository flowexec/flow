//nolint:gocognit,gocritic,nestif
package library

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/flowexec/tuikit/themes"
	"github.com/jahvon/glamour"

	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

const (
	// widthPadding is used when determining the width of the panes.
	// It is used to account for the spacing on left/right spacing of the panes
	widthPadding = 4

	allWorkspacesLabel    = "all workspaces"
	withoutNamespaceLabel = "w/o namespace"
	allNamespacesLabel    = "all namespaces"

	containerHelp        = "[ tab: split view ] [ ↑/↓: navigate pane ] [ ←/→: change pane ]"
	paneZeroHelp         = "[ o: open ] [ e: edit ] [ s:set ] ● [ space: show namespaces ]"
	paneZeroExpandedHelp = "[ o: open ] [ e: edit ] [ s:set ] ● [ space: hide namespaces ]"
	paneOneHelp          = "[ r: run ] [ e: edit ] [ c: copy ref ]"
	paneTwoHelp          = "[ r: run ] [ e: edit ] [ c: copy ref ]  ● [ f: change format ]"
)

var (
	// heightPadding is used when determining the height of the panes.
	// It is used to account for the header and footer
	heightPadding = themes.HeaderHeight + themes.FooterHeight
)

func (l *Library) View() string {
	l.paneZeroViewport.Style = paneStyle(0, l.theme, l.splitView)
	l.paneZeroViewport.SetContent(l.paneZeroContent())
	l.paneZeroViewport.SetYOffset(int(l.currentWorkspace + l.currentNamespace))

	l.paneOneViewport.Style = paneStyle(1, l.theme, l.splitView)
	l.paneOneViewport.SetContent(l.paneOneContent())

	l.paneTwoViewport.Style = paneStyle(2, l.theme, l.splitView)
	l.paneTwoViewport.SetContent(l.paneTwoContent())
	v := ctxVal(l.ctx.CurrentWorkspace.AssignedName(), l.ctx.Config.CurrentNamespace)
	header := l.theme.RenderHeader(appName, "ctx", v, l.termWidth)
	var panes string
	if l.splitView {
		panes = lipgloss.JoinHorizontal(
			lipgloss.Top,
			l.paneZeroViewport.View(),
			l.paneOneViewport.View(),
			l.paneTwoViewport.View(),
		)
	} else {
		switch l.currentPane {
		case 0, 1:
			panes = lipgloss.JoinHorizontal(
				lipgloss.Top,
				l.paneZeroViewport.View(),
				l.paneOneViewport.View(),
			)
		case 2:
			panes = l.paneTwoViewport.View()
		}
	}

	footer := l.footerContent()

	return lipgloss.JoinVertical(lipgloss.Top, header, panes, footer)
}

func (l *Library) SetNotice(notice string, level themes.OutputLevel) {
	if level == "" {
		level = themes.OutputLevelInfo
	}
	l.noticeText = l.theme.RenderLevel(notice, level)
}

func (l *Library) setSize() {
	l.termWidth = l.ctx.TUIContainer.Width()
	l.termHeight = l.ctx.TUIContainer.Height()
	p0, p1, p2 := calculateViewportWidths(l.termWidth-widthPadding, l.splitView)
	l.paneZeroViewport.Width = p0
	l.paneOneViewport.Width = p1
	l.paneTwoViewport.Width = p2
	l.paneZeroViewport.Height = l.termHeight - heightPadding
	l.paneOneViewport.Height = l.termHeight - heightPadding
	l.paneTwoViewport.Height = l.termHeight - heightPadding
}

func (l *Library) paneZeroContent() string {
	if !l.splitView && l.currentPane == 2 {
		return ""
	}

	var sb strings.Builder
	l.mu.RLock()
	workspaces := l.visibleWorkspaces
	namespaces := l.visibleNamespaces
	l.mu.RUnlock()

	sb.WriteString(renderPaneTitle("Workspaces", len(workspaces), l.currentPane == 0, l.theme))

	numWs := len(workspaces)
	numNs := len(namespaces)
	if numWs == 0 {
		sb.WriteString(l.theme.RenderError("No workspaces found"))
		return sb.String()
	}
	paneWidth, _, _ := calculateViewportWidths(l.termWidth, l.splitView)

	for i, ws := range workspaces {
		prefix := "◌ "
		if uint(i) == l.currentWorkspace && !l.showNamespaces {
			prefix = "● "
		} else if uint(i) == l.currentWorkspace && l.showNamespaces {
			prefix = "◉ "
		}

		if uint(i) == l.currentWorkspace {
			sb.WriteString(renderSelection(prefix+truncateText(ws, paneWidth), l.theme))
			sb.WriteString("\n")
			if numNs == 1 {
				sb.WriteString(renderDescription(fmt.Sprintf("  %d namespace", numNs), l.theme))
			} else {
				sb.WriteString(renderDescription(fmt.Sprintf("  %d namespaces", numNs), l.theme))
			}
			sb.WriteString("\n")

			if l.showNamespaces {
				for j, ns := range namespaces {
					if uint(j) == l.currentNamespace {
						sb.WriteString(renderSecondarySelection("  > "+truncateText(ns, paneWidth), l.theme))
					} else {
						sb.WriteString(renderInactive("    "+truncateText(ns, paneWidth), l.theme))
					}
					sb.WriteString("\n")
				}
			}
		} else {
			sb.WriteString(renderInactive(prefix+truncateText(ws, paneWidth), l.theme))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (l *Library) paneOneContent() string {
	if !l.splitView && l.currentPane == 2 {
		return ""
	}

	var sb strings.Builder
	l.mu.RLock()
	sb.WriteString(renderPaneTitle("Executables", len(l.visibleExecutables), l.currentPane == 1, l.theme))
	if len(l.visibleExecutables) == 0 {
		l.mu.RUnlock()
		sb.WriteString(l.theme.RenderError("No executables found"))
		return sb.String()
	}

	_, paneWidth, _ := calculateViewportWidths(l.termWidth, l.splitView)

	curWs := l.visibleWorkspaces[l.currentWorkspace]
	var curNs string
	if len(l.visibleNamespaces) > 0 {
		curNs = l.visibleNamespaces[l.currentNamespace]
	}
	visibleExecutables := l.visibleExecutables
	l.mu.RUnlock()

	for i, ex := range visibleExecutables {
		if uint(i) == l.currentExecutable {
			indicator := "*"
			if (l.ctx.CurrentWorkspace != nil && ex.Workspace() == l.ctx.CurrentWorkspace.AssignedName()) ||
				(ex.Visibility != nil && *ex.Visibility == executable.ExecutableVisibility(common.VisibilityPublic)) {
				// indicate if runnable from the current ctx
				indicator = "▶"
			}
			refStr := indicator + " " + truncateText(shortRef(ex.Ref(), curWs, curNs), paneWidth)
			sb.WriteString(renderSelection(refStr, l.theme))
		} else {
			sb.WriteString(renderInactive("  "+truncateText(shortRef(ex.Ref(), curWs, curNs), paneWidth), l.theme))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (l *Library) paneTwoContent() string {
	l.mu.RLock()
	if len(l.visibleExecutables) == 0 {
		l.mu.RUnlock()
		return ""
	} else if !l.splitView && l.currentPane != 2 {
		l.mu.RUnlock()
		return ""
	}

	ex := l.visibleExecutables[l.currentExecutable]
	l.mu.RUnlock()

	_, _, maxWidth := calculateViewportWidths(l.termWidth, l.splitView)
	paneTwoMaxWidth := math.Floor(float64(maxWidth) * 0.95)
	mdStyles, err := l.theme.GlamourMarkdownStyleJSON()
	if err != nil {
		return l.theme.RenderError(fmt.Sprintf("unable to render markdown: %s", err.Error()))
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(mdStyles)),
		glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(int(paneTwoMaxWidth)),
	)
	if err != nil {
		return l.theme.RenderError(fmt.Sprintf("unable to render markdown: %s", err.Error()))
	}

	content := ex.Markdown()
	switch l.currentFormat {
	case 0:
		content = ex.Markdown()
	case 1:
		content, err = ex.YAML()
		if err != nil {
			return l.theme.RenderError(fmt.Sprintf("unable to render yaml: %s", err.Error()))
		}
		content = fmt.Sprintf("```yaml\n%s\n```", content)
	case 2:
		content, err = ex.JSON()
		if err != nil {
			return l.theme.RenderError(fmt.Sprintf("unable to render json: %s", err.Error()))
		}
		content = fmt.Sprintf("```json\n%s\n```", content)
	}
	viewStr, err := renderer.Render(content)
	if err != nil {
		return l.theme.RenderError(fmt.Sprintf("unable to render markdown: %s", err.Error()))
	}

	return viewStr
}

func (l *Library) footerContent() string {
	help := l.showHelp
	if help && l.currentHelpPage != 0 {
		return l.theme.RenderFooter(fmt.Sprintf("2/2 %s ● %s", "[ h: exit help ]", containerHelp), l.termWidth)
	}

	footerPrefix := "[ q/ctrl+c: quit] [ h: help ]"
	if help {
		footerPrefix = "1/2 [ h: show more ]"
	}
	switch l.currentPane {
	case 0:
		if help && l.showNamespaces {
			return l.theme.RenderFooter(
				fmt.Sprintf("%s ● %s", footerPrefix, paneZeroExpandedHelp), l.termWidth,
			)
		} else if help {
			return l.theme.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, paneZeroHelp), l.termWidth)
		} else {
			l.mu.RLock()
			if l.currentWorkspace < uint(len(l.visibleWorkspaces)) {
				ws := l.visibleWorkspaces[l.currentWorkspace]
				l.mu.RUnlock()
				if ws == allWorkspacesLabel {
					break
				}
				var wsCfg *workspace.Workspace
				for i, w := range l.allWorkspaces {
					if w.AssignedName() == ws {
						wsCfg = l.allWorkspaces[i]
					}
				}
				if wsCfg == nil {
					l.ctx.Logger.Errorf("unable to find workspace config for %s", ws)
					break
				}

				var info string
				switch {
				case l.noticeText != "":
					info = l.noticeText
				case len(wsCfg.Tags) > 0:
					info = fmt.Sprintf("%s(%s) -> %s", wsCfg.DisplayName, common.Tags(wsCfg.Tags).PreviewString(), wsCfg.Location())
				default:
					info = fmt.Sprintf("%s -> %s", wsCfg.DisplayName, wsCfg.Location())
				}
				return l.theme.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, info), l.termWidth)
			} else {
				l.mu.RUnlock()
			}
		}
	case 1, 2:
		if help {
			helpStr := paneOneHelp
			if l.currentPane == 2 {
				helpStr = paneTwoHelp
			}

			return l.theme.RenderFooter(
				fmt.Sprintf("%s ● %s", footerPrefix, helpStr), l.termWidth,
			)
		} else {
			l.mu.RLock()
			if l.currentExecutable < uint(len(l.visibleExecutables)) {
				var info string
				switch {
				case l.noticeText != "":
					info = l.noticeText
				default:
					exec := l.visibleExecutables[l.currentExecutable]
					l.mu.RUnlock()
					info = exec.FlowFilePath()
				}
				return l.theme.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, info), l.termWidth)
			} else {
				l.mu.RUnlock()
			}
		}
	}
	return l.theme.RenderFooter(footerPrefix, l.termWidth)
}
