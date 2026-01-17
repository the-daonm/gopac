# gopac

[![gopac](https://img.shields.io/static/v1?label=gopac&message=v1.3.0&color=1793d1&style=for-the-badge&logo=archlinux)](https://aur.archlinux.org/packages/gopac/)
[![gopac-bin](https://img.shields.io/static/v1?label=gopac-bin&message=v1.3.0&color=1793d1&style=for-the-badge&logo=archlinux)](https://aur.archlinux.org/packages/gopac-bin/)

A warm, beautiful TUI for Arch Linux package management, written in Go.
**gopac** allows you to search, view, and install packages from both the official Arch repositories and the AUR simultaneously.

> [!NOTE]
> This application uses Nerd Font icons. For the best experience, please use a [Nerd Font](https://www.nerdfonts.com/) in your terminal.

![Screenshot](screenshot.png)

## Features

- **Unified Search**: Search Official repos and AUR at the same time.
- **Smart Sorting**: Exact matches and installed packages appear first.
- **Beautiful UI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) using a cozy Gruvbox theme.
- **Detailed Views**: View maintainer info, votes, versions, and more.
- **Fast**: Written in Go for speed.

## Installation

### From AUR

```bash
yay -S gopac # paru -S gopac
```

### Manual Build

```bash
git clone https://github.com/the-daonm/gopac.git
cd gopac
go build
sudo mv gopac /usr/bin/
mkdir -p ~/.config/fish/completions
cp completions/gopac.fish ~/.config/fish/completions/
```

## Usage

Run the app:

```bash
gopac
```

## Configuration

**gopac** looks for a configuration file at `~/.config/gopac/config.yaml`.

Example configuration:

```yaml
aur_helper: yay
theme: dracula
```

### Available Themes
- `gruvbox` (default)
- `onedark`
- `dracula`
- `nord`
- `catppuccin`

### AUR Helper

**gopac** automatically detects your AUR helper. It checks for the following tools in this order:

1. `paru`
2. `yay`
3. `pikaur`
4. `aura`
5. `trizen`

## License

MIT
