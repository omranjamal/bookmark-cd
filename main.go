package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jwalton/go-supportscolor"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/muesli/termenv"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var version = "vvvv"

var helpText = `

# BOOKMARK-CD

> CLI utility that lets you interactively pick the bookmarked
  directory you want to cd into.

Usage:
  bcd                  Start interactive mode bookmark picker
  bcd [SEARCH_TERM]    Search by SEARCH_TERM and automatically cd into it, if there is only one match

  [UP] / [DOWN] Arrow Keys (interactive mode only):
    Lets you choose a bookmarked directory from the list

  Start Typing (interactive mode only):
    This will allow you to filter the suggested bookmarked directories

Flags:
  -h / --help                  Show this help message

  --shell [ALIAS]              Show the shell function code with name set to ALIAS (optional)
                               Mostly useful for manual installation.

  --install FILE [ALIAS]       Add or update the shell function in a shell startup file
                               like ~/.bashrc; Setting an ALIAS will change the function
                               name from bcd to ALIAS
`

var shellFunction string = `# start: bookmark-cd
bcd() {
  TARGETPATH=$(bookmark-cd $1)

  if [ ! -z "${TARGETPATH}" ] ; then
    cd "${TARGETPATH}"
  fi
}
# end: bookmark-cd`

type Candidate struct {
	name string
	path string
}

type FilteredCandidate struct {
	candidate Candidate
	rank      int
}

type State struct {
	selected           bool
	cursor             int
	exited             bool
	filteredCandidates *[]FilteredCandidate
}

type model struct {
	searchInput textinput.Model
	searchText  string
	candidates  []Candidate // bookmarks
	state       *State
}

func getCandidates() []Candidate {
	filename := os.Getenv("HOME") + "/.config/gtk-3.0/bookmarks"
	file, err := os.Open(filename)

	if err != nil {
		fmt.Printf("could not open file %s \n", filename)
	}

	scanner := bufio.NewScanner(file)
	candidates := make([]Candidate, 0, 32)

	for scanner.Scan() {
		line := scanner.Text()
		chunks := strings.SplitN(line, " ", 2)

		pathUrl, _ := url.QueryUnescape(chunks[0])

		candidates = append(candidates, Candidate{
			name: (func() string {
				if len(chunks) > 1 {
					return chunks[1]
				} else {
					basename, _ := url.QueryUnescape(path.Base(chunks[0]))
					return basename
				}
			})(),
			path: strings.Replace(pathUrl, "file://", "", 1),
		})
	}

	return candidates
}

func getFilteredCandidates(candidates *[]Candidate, searchText string) *[]FilteredCandidate {
	filteredCandidates := make([]FilteredCandidate, 0, 32)

	if searchText == "" {
		for _, c := range *candidates {
			filteredCandidates = append(filteredCandidates, FilteredCandidate{
				candidate: c,
				rank:      1,
			})
		}
	} else {
		lowerSearchText := strings.ToLower(searchText)

		for _, c := range *candidates {
			rank := fuzzy.RankMatch(
				lowerSearchText,
				strings.ToLower(c.name),
			)

			if rank >= 0 {
				filteredCandidates = append(filteredCandidates, FilteredCandidate{
					candidate: c,
					rank:      rank,
				})
			}
		}

		sort.Slice(filteredCandidates, func(i, j int) bool {
			return filteredCandidates[i].rank > filteredCandidates[j].rank
		})
	}

	return &filteredCandidates
}

func initModel(initialSearchText string) model {
	searchInput := textinput.New()

	searchInput.Placeholder = "Search"
	searchInput.Focus()
	searchInput.CharLimit = 156
	searchInput.Width = 32
	searchInput.Prompt = ": "

	if initialSearchText != "" {
		searchInput.SetValue(initialSearchText)
	}

	candidates := getCandidates()

	return model{
		searchInput: searchInput,
		searchText:  "",
		candidates:  candidates,

		state: &State{
			selected:           false,
			filteredCandidates: getFilteredCandidates(&candidates, initialSearchText),
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

var nameStyle = lipgloss.NewStyle().Bold(true)
var pathStyle = lipgloss.NewStyle().Faint(true)

func (m model) candidatesView() string {
	buf := bytes.NewBufferString("")

	l := len(m.candidates)
	current := m.state.cursor % l

	for i, fc := range *m.state.filteredCandidates {
		name := nameStyle.Render(fc.candidate.name)

		bookmarkPath := pathStyle.Render(
			strings.Replace(fc.candidate.path, os.Getenv("HOME"), "~", 1),
		)

		row := fmt.Sprintf("%s %s\n", name, bookmarkPath)

		if i == current {
			buf.WriteString("> " + row)
		} else {
			buf.WriteString("  " + row)
		}

	}

	return buf.String()
}

func (m model) View() string {
	if m.state.exited {
		return ""
	} else {
		return fmt.Sprintf(
			"%s\n%s",
			m.searchInput.View(),
			m.candidatesView(),
		)
	}
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch message := message.(type) {
	case tea.KeyMsg:
		switch message.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			m.state.exited = true

			if message.Type == tea.KeyEnter {
				m.state.selected = true
			}

			return m, tea.Quit

		case tea.KeyDown:
			m.state.cursor++

			if m.state.cursor >= len(m.candidates) {
				m.state.cursor = 0
			}

			return m, nil

		case tea.KeyUp:
			if m.state.cursor > 0 {
				m.state.cursor--
			} else {
				m.state.cursor = len(*m.state.filteredCandidates) - 1
			}

			return m, nil
		}
	}

	m.searchInput, cmd = m.searchInput.Update(message)
	searchText := m.searchInput.Value()

	if m.searchText != searchText {
		m.state.filteredCandidates = getFilteredCandidates(&m.candidates, searchText)
		m.searchText = searchText
		m.state.cursor = 0
	}

	return m, cmd
}

func setupTerminal() {
	term := supportscolor.Stderr()

	if term.Has16m {
		lipgloss.SetColorProfile(termenv.TrueColor)
	} else if term.Has256 {
		lipgloss.SetColorProfile(termenv.ANSI256)
	} else {
		lipgloss.SetColorProfile(termenv.ANSI)
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	return err
}

func install(filePath string, alias string) {
	fileDir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	backupFileName := fmt.Sprintf("%s.%d.bcd-install-backup", fileName, time.Now().UnixMilli())
	backupFilePath := filepath.Join(fileDir, backupFileName)

	copyError := copyFile(filePath, backupFilePath)

	if copyError != nil {
		log.Fatal(copyError)
	}

	b, readError := os.ReadFile(filePath)

	if readError != nil {
		log.Fatal(readError)
	}

	fileContents := string(b)
	lines := strings.Split(strings.TrimSpace(fileContents), "\n")
	modifiedLines := make([]string, 0, len(lines)+128)

	started := false
	ended := false

	for _, line := range lines {
		if !started && strings.TrimSpace(line) == "# start: bookmark-cd" {
			started = true
			continue
		} else if started && strings.TrimSpace(line) == "# end: bookmark-cd" {
			ended = true
			continue
		}

		if !started || ended {
			modifiedLines = append(modifiedLines, line)
		}
	}

	shellFunctionLines := strings.Split(
		strings.Replace(shellFunction, "bcd", alias, 1),
		"\n",
	)

	modifiedLines = append(
		modifiedLines,
		shellFunctionLines...,
	)

	writeError := os.WriteFile(
		filePath,
		[]byte(
			strings.TrimSpace(strings.Join(modifiedLines, "\n"))+"\n",
		),
		0644,
	)

	if writeError != nil {
		log.Fatal(writeError)
	}

	backupRemovalError := os.Remove(backupFilePath)

	if backupRemovalError != nil {
		log.Fatal(backupRemovalError)
	}
}

func main() {
	setupTerminal()

	isPrintShellScript := 0
	isInstall := 0

	alias := "bcd"
	shellFile := ""

	search := make([]string, 0, 8)

	for _, arg := range os.Args[1:] {
		if arg == "--shell" {
			isPrintShellScript = 1
		} else if arg == "--install" {
			isInstall = 1
		} else if arg == "--version" || arg == "-v" {
			fmt.Println("bookmark-cd " + version)
			return
		} else if arg == "--help" || arg == "-h" {
			fmt.Print(helpText)
			return
		} else {
			if isPrintShellScript == 1 {
				alias = arg
			} else if isInstall == 1 {
				if shellFile != "" {
					alias = arg
				} else {
					absolutePath, absolutePathError := filepath.Abs(arg)

					if absolutePathError != nil {
						log.Fatal(absolutePathError)
					}

					shellFile = absolutePath
				}
			} else {
				search = append(search, arg)
			}
		}
	}

	if (isPrintShellScript + isInstall) > 1 {
		fmt.Println("ERROR: can't use --shell and --install together")
		os.Exit(1)
		return
	}

	if isInstall == 1 {
		if shellFile == "" {
			fmt.Println("ERROR: must provide a shell file to modify")
			os.Exit(1)
			return
		} else {
			install(shellFile, alias)
			return
		}
	}

	initialSearchText := strings.Join(search, " ")

	if isPrintShellScript == 1 {
		fmt.Fprintf(
			os.Stdout,
			"%s\n",
			strings.Replace(
				shellFunction,
				"bcd",
				alias,
				1,
			),
		)

		return
	}

	m := initModel(initialSearchText)

	if len(*m.state.filteredCandidates) == 1 {
		os.Stdout.WriteString((*m.state.filteredCandidates)[0].candidate.path)
		return
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	_, err := p.Run()

	if err != nil {
		log.Fatal(err)
	}

	if m.state.selected {
		if len(*m.state.filteredCandidates) > 0 {
			targetPath := (*m.state.filteredCandidates)[m.state.cursor].candidate.path
			os.Stdout.WriteString(targetPath)
		}
	}
}
