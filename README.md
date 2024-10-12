# bookmark-cd

> The fastest way to `cd` into your GTK bookmarks.

## Installation

```bash
git clone git@github.com:omranjamal/bookmark-cd.git
cd ./bookmark-cd

go get
go build

sudo ln -s ./bookmark-cd /usr/bin

# assuming you're using bash
bookmark-cd --shell --eval >> ~/.bashrc
```

## Development

```bash
git clone git@github.com:omranjamal/bookmark-cd.git

cd ./bookmark-cd

go get
go run main.go
```

## License

MIT