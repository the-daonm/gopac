# Fish completion for gopac

# Disable file completions unless specifically needed
complete -c gopac -f

# Helper flag
complete -c gopac -s H -l helper -d 'Specify AUR helper to use' -ra 'paru yay pikaur aura trizen'

# Theme flag
complete -c gopac -s t -l theme -d 'Specify UI theme' -ra 'gruvbox onedark dracula nord catppuccin'

# Version flag
complete -c gopac -s v -l version -d 'Show version information'

# Help flag
complete -c gopac -s h -l help -d 'Show help'
