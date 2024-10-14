package bcd

var ShellFunction string = `# start: bookmark-cd
bcd() {
  TARGETPATH=$("$HOME/.local/share/omranjamal/bookmark-cd/bookmark-cd" $1)

  if [ ! -z "${TARGETPATH}" ] ; then
    cd "${TARGETPATH}"
  fi
}
# end: bookmark-cd`
