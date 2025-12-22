# Fish completion for gopac

# Disable file completions unless specifically needed
complete -c gopac -f

# Helper flag
complete -c gopac -s H -l helper -d 'Specify AUR helper to use' -ra 'paru yay pikaur aura trizen'

# Help flag
complete -c gopac -s h -l help -d 'Show help'

# Example for AUR_HELPER environment variable completion
# Note: Fish doesn't natively support per-command environment variable completions 
# in the 'VAR=val cmd' syntax without extra plugins, but we can provide a general
# helper for setting it.
