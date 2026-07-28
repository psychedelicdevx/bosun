Output docs/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"

Hide
Type "go build -o bosun ./cmd" Enter
Sleep 2s
Type "./bosun --demo" Enter
Sleep 2s
Show

Sleep 1s

# compose groups: collapse and expand the shop stack
Space
Sleep 1s
Space
Sleep 800ms

# move the highlight down onto a container
Down@450ms 2
Sleep 1s

# stream its logs, focus the pane, scroll
Enter
Sleep 3s
Tab
Sleep 400ms
Down@350ms 4
Sleep 1s
Escape
Sleep 800ms

# live stats
Type "S"
Sleep 3s
Escape
Sleep 800ms

# filter the list by name
Type "/"
Sleep 500ms
Type "blog"
Sleep 1800ms
Escape
Sleep 800ms

# help overlay
Type "?"
Sleep 2500ms
Type "?"
Sleep 800ms

Type "q"
Sleep 1s
