# bookmark-cd

![GitHub Release Date](https://img.shields.io/github/release-date/omranjamal/bookmark-cd)
![GitHub Release](https://img.shields.io/github/v/release/omranjamal/bookmark-cd)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues/omranjamal/bookmark-cd)


> The fastest way to `cd` into your folder bookmarks that you create on your file manager.

![DEMO](https://raw.githubusercontent.com/omranjamal/bookmark-cd/refs/heads/static/demo.gif)

## Features

1. Interactive
2. Fuzzy search
3. Non interactive mode for even quicker `cd`
4. Works with:
   5. [nautilus](https://apps.gnome.org/Nautilus/)
   6. [thunar](https://docs.xfce.org/xfce/thunar/start)
   7. [nemo](https://github.com/linuxmint/nemo)
   8. [caja](https://wiki.mate-desktop.org/mate-desktop/applications/caja/)

## Usage

```bash
# interactive mode:
bcd

# non interactive `cd` if only one match is present.
bcd [search]
```

- `Up` / `Down` to select a target bookmark.
- Starting typing to filter list of bookmarks.

## Installation
Make sure you have `curl` installed on your system

```bash
curl -sL https://github.com/omranjamal/bookmark-cd/releases/latest/download/install.sh -o - | sh -
```

### Manual Installation

```bash
# Step 1: Download the binary (check releases page for all available binaries)
curl -L -o bookmark-cd https://github.com/omranjamal/bookmark-cd/releases/latest/download/bookmark-cd_v1.0.2_amd64

# Step 2: Add execution permissions
chmod +x ./bookmark-cd

# Step 3: Move into a directory that is on your path.
mv ./bookmark-cd /usr/bin/bookmark-cd

# Step 4: Add to shell
# ⚠️⚠️ WARNING: make sure you wrote `>>` and not `>` ⚠️⚠️
bookmark-cd --shell >> ~/.bashrc
```

### Setting Different Alias

You can either change the function name in your
`~/.bashrc` / `~/.zshrc` file from `bcd` to something
else.

OR, you could add the alias in Step 4 from above by passing
as the last argument.

```bash
# ⚠️⚠️ WARNING: make sure you wrote `>>` and not `>` ⚠️⚠️
bookmark-cd --shell bookcd >> ~/.bashrc
```

`bookcd` being the different alias that you want.

## Development

```bash
git clone git@github.com:omranjamal/bookmark-cd.git

cd ./bookmark-cd

go get
go run main.go
```

## License

MIT