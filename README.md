<p align="center">
  <img src="assets/logo.svg" width="128" height="128" alt="gopac logo" />
</p>

<h1 align="center">gopac</h1>

<p align="center">
  <strong>A warm, keyboard-driven TUI for Arch Linux package management.</strong><br>
  Search the official repos and the AUR together, review before you install, and keep your system tidy.
</p>

<p align="center">
  <a href="https://aur.archlinux.org/packages/gopac/"><img src="https://img.shields.io/aur/version/gopac?style=flat-square&color=1793d1&label=gopac" alt="AUR version" /></a>
  <a href="https://aur.archlinux.org/packages/gopac-bin/"><img src="https://img.shields.io/aur/version/gopac-bin?style=flat-square&color=1793d1&label=gopac-bin" alt="AUR binary version" /></a>
  <a href="https://github.com/the-daonm/gopac/actions/workflows/ci.yaml"><img src="https://img.shields.io/github/actions/workflow/status/the-daonm/gopac/ci.yaml?branch=main&style=flat-square&label=CI" alt="CI status" /></a>
  <a href="LICENCE"><img src="https://img.shields.io/static/v1?label=license&message=MIT&color=ea6962&style=flat-square" alt="MIT License" /></a>
</p>

<p align="center">
  <img src="demo.gif" width="100%" style="max-width: 800px; border-radius: 8px;" alt="gopac demo" />
</p>

---

## Highlights

- **One search, two sources.** Official repositories and the AUR are searched in parallel and merged into a single list.
- **Your system at a glance.** Open gopac with an empty search to browse every installed package.
- **Pending updates.** The **UPDATES** tab shows repo and AUR upgrades as `old → new` before you run a full upgrade.
- **Orphan cleanup.** The **ORPHANS** tab lists dependencies nothing needs anymore; queue them all with `a`, apply with `I`.
- **Review before you trust.** Read any AUR package's **PKGBUILD** in place. Out-of-date and unmaintained AUR packages are flagged.
- **Batch changes safely.** Queue installs and removals, then confirm them all in one summary screen.
- **Looks good in your terminal.** Five built-in themes, mouse support, and a layout that adapts to narrow windows.

---

## Installation

### From the AUR

```bash
paru -S gopac        # build from source
paru -S gopac-bin    # pre-built binary
```

(`yay -S ...` works just as well.)

### From source

Requires Go 1.25 or newer.

```bash
git clone https://github.com/the-daonm/gopac.git
cd gopac
make build
sudo install -Dm755 gopac /usr/bin/gopac
```

### Requirements

| Dependency | Required? | Why |
| :--- | :--- | :--- |
| `pacman` | Yes | Searching, installing and removing packages |
| An AUR helper (`paru`, `yay`, ...) | For AUR installs | gopac hands AUR builds to your helper |
| `pacman-contrib` | Recommended | Provides `checkupdates`, so the UPDATES tab sees fresh data without root. Without it, gopac falls back to `pacman -Qu`, which only knows about your last `pacman -Sy` |
| A [Nerd Font](https://www.nerdfonts.com/) | Recommended | Icons in the search bar and queue markers |

---

## Usage

Run `gopac`. You start in the search bar with your installed packages listed below it. Type to search; results appear as you type.

### Tabs

Switch tabs with `h` / `l` (or the arrow keys, or a mouse click). Each tab shows how many packages it contains.

| Tab | Shows |
| :--- | :--- |
| **ALL** | Every result: search results, or all installed packages when the search is empty |
| **AUR** | Results from the AUR (or installed packages not in any repo) |
| **OFFICIAL** | Results from the official repositories |
| **INSTALLED** | Results that are already installed |
| **UPDATES** | Installed packages with a newer version available |
| **ORPHANS** | Packages installed as dependencies that nothing requires anymore |

Search results are ordered by exact name match first, then official packages before AUR ones, then AUR packages by votes.

### Common tasks

| I want to... | Do this |
| :--- | :--- |
| Install a package | Search, select it, press `enter`, confirm with `y` |
| Install several at once | Press `space` on each, then `I` to review and apply |
| Remove a package | Select an installed package, press `enter` |
| Check an AUR package before installing | Select it and press `p` to read its PKGBUILD |
| Upgrade everything | Press `U` (or `enter` in the UPDATES tab) |
| Clean up orphans | Open ORPHANS, press `a` to queue all, then `I` |

---

## Keybindings

Press `?` at any time for the in-app help.

**General**

| Key | Action |
| :--- | :--- |
| `/` | Focus the search bar |
| `tab` / `shift+tab` | Cycle focus: search → list → details |
| `U` | Upgrade the whole system |
| `r` | Refresh installed packages, updates and the current search |
| `?` | Toggle help |
| `ctrl+c` | Quit |

**Search bar**

| Key | Action |
| :--- | :--- |
| `enter` | Search now and jump to the list |
| `↑` / `↓` | Browse search history |
| `esc` | Leave the search bar |

**Package list**

| Key | Action |
| :--- | :--- |
| `j` / `k`, `↑` / `↓` | Move the selection |
| `h` / `l`, `←` / `→` | Switch tab |
| `enter` | Install or remove the selected package (upgrade all in UPDATES) |
| `space` | Queue / unqueue the selected package |
| `a` | Queue / unqueue every visible package |
| `I` | Review and apply the queue |
| `C` | Clear the queue |
| `p` | Toggle the PKGBUILD view (AUR packages) |

**Details pane**

| Key | Action |
| :--- | :--- |
| `j` / `k`, `↑` / `↓` | Scroll one line |
| `ctrl+d` / `ctrl+u` | Scroll ten lines |
| `esc` | Back to the list |

The mouse works too: click a tab, a panel or a package, and scroll whichever panel is under the pointer.

---

## Configuration

gopac reads `~/.config/gopac/config.yaml` (or `$XDG_CONFIG_HOME/gopac/config.yaml`). Every setting is optional.

```yaml
# AUR helper to use for AUR installs and system upgrades.
# Leave unset to auto-detect.
aur_helper: paru

# UI theme: gruvbox (default), onedark, dracula, nord, catppuccin
theme: gruvbox
```

If the file can't be parsed, gopac prints a warning and starts with defaults.

### Command-line flags

```text
Usage: gopac [flags]

  -H, --helper     AUR helper to use (e.g. paru, yay)
  -t, --theme      UI theme (gruvbox, onedark, dracula, nord, catppuccin)
  -v, --version    Show version information
```

```bash
gopac -t dracula       # try a theme without editing the config
gopac --helper yay     # use yay for this session
```

### Choosing the AUR helper

gopac picks the first match from:

1. the `--helper` flag
2. `aur_helper` in the config file
3. the `AUR_HELPER` environment variable (if that program is on your `$PATH`)
4. the first installed of `paru`, `yay`, `pikaur`, `aura`, `trizen`

Without a helper, gopac still searches the AUR and manages official packages, but refuses AUR installs and tells you why.

### Themes

| Theme | Value | Style |
| :--- | :--- | :--- |
| **Gruvbox** (default) | `gruvbox` | Retro, warm and cozy |
| **OneDark** | `onedark` | Balanced Atom-style dark |
| **Dracula** | `dracula` | High-contrast purple and pink |
| **Nord** | `nord` | Cool arctic blues |
| **Catppuccin** | `catppuccin` | Soft mocha pastels |

### Shell completions

Fish completions ship with the AUR packages. For a source install:

```bash
install -Dm644 completions/gopac.fish ~/.config/fish/completions/gopac.fish
```

---

## How gopac changes your system

gopac never edits package state itself; every change runs a command you could type yourself, in your terminal, with its normal prompts:

| Action | Command |
| :--- | :--- |
| Install official packages | `sudo pacman -S -- <pkgs>` |
| Install AUR packages | `<helper> -S -- <pkgs>` |
| Remove packages | `sudo pacman -Rns -- <pkgs>` |
| Upgrade the system | `<helper> -Syu` or `sudo pacman -Syu` |

If a command fails, gopac keeps its output on screen until you press Enter, and keeps your queue so you can retry.

Updates are only offered as a full system upgrade, because [partial upgrades are unsupported](https://wiki.archlinux.org/title/System_maintenance#Partial_upgrades_are_unsupported) on Arch.

**AUR packages are user-submitted.** gopac shows a warning before installing them; read the PKGBUILD (`p`) first.

### Files

| Path | Contents |
| :--- | :--- |
| `~/.config/gopac/config.yaml` | Your configuration |
| `~/.cache/gopac/history` | The last 100 searches |

---

## Development

```bash
make build    # build ./gopac with the version from the latest git tag
make test     # go test ./...
```

Pushing a `v*` tag runs the release workflow, which publishes the GitHub release and updates both AUR packages.

Bug reports and pull requests are welcome at [github.com/the-daonm/gopac](https://github.com/the-daonm/gopac/issues).

---

## License

MIT. See [LICENCE](LICENCE).
