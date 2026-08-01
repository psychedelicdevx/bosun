Output docs/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1160
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
Sleep 800ms
Space
Sleep 800ms

# move onto a running container and stream its logs
Down@400ms 2
Sleep 800ms
Enter
Sleep 2500ms

# filter the log stream live
Type "/"
Sleep 400ms
Type "job"
Sleep 2s
Escape
Sleep 500ms

# copy the shown logs to the clipboard
Type "y"
Sleep 1800ms
Escape
Sleep 800ms

# tour the tabs: images, volumes, networks
Type "2"
Sleep 1600ms
Type "3"
Sleep 1400ms
Down@400ms 2
Sleep 800ms
Type "4"
Sleep 1600ms
Down@400ms 1
Sleep 1200ms

# back to containers, operate a whole stack from its header
Type "1"
Sleep 500ms
Up@400ms 2
Sleep 700ms
Type "x"
Sleep 2200ms
Type "s"
Sleep 2200ms

# keybindings overlay
Type "?"
Sleep 2500ms
Type "?"
Sleep 700ms

Type "q"
Sleep 1s
