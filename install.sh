#!/usr/bin/env sh

V="${VERSION:-vvvv}"
REPORTED_ARCH="$(uname -m)"

if [ "$REPORTED_ARCH" = "x86_64" ]; then
  DOWNLOAD_ARCH="amd64"
elif [ "$REPORTED_ARCH" = "i686" ]; then
  DOWNLOAD_ARCH="386"
elif [ "$REPORTED_ARCH" = "i386" ]; then
  DOWNLOAD_ARCH="386"
elif [ "$REPORTED_ARCH" = "arm" ]; then
  DOWNLOAD_ARCH="arm"
elif [ "$REPORTED_ARCH" = "armv7l" ]; then
  DOWNLOAD_ARCH="arm"
elif [ "$REPORTED_ARCH" = "aarch64" ]; then
  DOWNLOAD_ARCH="arm64"
else
  echo "Unknown Architecture"
  exit
fi

URL="https://github.com/omranjamal/bookmark-cd/releases/latest/download/bookmark-cd_${V}_${DOWNLOAD_ARCH}"
TARGET="${INSTALL_TO:-/usr/bin/bookmark-cd}"

download() {
  echo "> 📥 downloading $URL -> ./bookmark-cd" && \
      curl -s -L -o "./bookmark-cd" "$URL" && return 0
}

update_permissions() {
  echo "> 💪 setting execution permission" && chmod +x ./bookmark-cd && return 0
}

move() {
  echo "> 🚚 moving ./bookmark-cd -> $TARGET (👑 this will require root privileges)" && \
       sudo mv "./bookmark-cd" "$TARGET" && return 0
}

add_to_shell() {
  if [ -f "$HOME/.bashrc" ]; then
    echo "> 👉 Detected ~/.bashrc"
    echo "> ⚡ Adding to ~/.bashrc"
    bookmark-cd --shell >> "$HOME/.bashrc"

    echo "\e[0m"
    echo "    🚀 Run this, or re-start your bash terminal:"
    echo "       source ~/.bashrc"
    echo "\e[2m"
  fi

  if [ -f "$HOME/.zshrc" ]; then
    echo "> 👉 Detected ~/.zshrc"
    echo "> ⚡ Adding to ~/.zshrc"
    bookmark-cd --shell >> "$HOME/.zshrc"

    echo "\e[0m"
    echo "    🚀 Run this, or re-start your zsh terminal:"
    echo "       source ~/.zshrc"
    echo "\e[2m"
  fi
}

completed() {
  echo "> DONE. ✅"
}

printf "\e[2m"

download && \
  update_permissions && \
  move && \
  add_to_shell && \
  completed

printf "\e[0m"
