Output docs/demo/hero.gif
Output docs/demo/hero.mp4

Require flow

Set WindowBar Colorful
Set WindowBarSize 60
Set Shell "bash"
Set FontSize 20
Set Width 1400
Set Height 800
Set Padding 20
Set TypingSpeed 60ms
Set Theme "GitHub Dark"
Set PlaybackSpeed 0.75

Env DISABLE_FLOW_INTERACTIVE "false"

# ── HIDDEN SETUP ─────────────────────────────────────────────────────────────
Hide
Type "flow workspace add my-project ./docs/demo/project"
Enter
Sleep 2s
Type "flow sync"
Sleep 2s
Type "flow workspace switch my-project"
Enter
Sleep 2s
Type "clear"
Enter
Sleep 500ms
Show

# ── ACT 1: Quick-run from CLI ────────────────────────────────────────────────
Sleep 1s
Type "flow run demo:greet"
Enter
Sleep 1s
Type "Flow"
Enter
Sleep 1s
Type "pink"
Enter
Sleep 2s

# ── ACT 2: Enter the library ─────────────────────────────────────────────────
Type "flow browse"
Enter
Sleep 1s
Space

# Navigate down to my-project workspace in the left pane
Down@500ms 5
Sleep 1s

# Drill into executable list with Right arrow
Right@500ms
Sleep 1s

# Show the filter/search box
Type "/"
Sleep 500ms
Type "deploy"
Sleep 2s
Enter
Sleep 500ms

# View documentation for the executable
Enter
Sleep 1s

# Run it the deploy executable
Sleep 1s
Type "r"
# The TUI exits immediately; output runs in the shell.
Sleep 4s
