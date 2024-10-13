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
	"log"
	"os"
	"path"
	"sort"
	"strings"
)

var version = "vvvv"

var shellFunction string = `
bcd() {
  TARGETPATH=$(bookmark-cd $1)

  if [ ! -z "${TARGETPATH}" ] ; then
    cd $TARGETPATH
  fi
}
`

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

		candidates = append(candidates, Candidate{
			name: (func() string {
				if len(chunks) > 1 {
					return chunks[1]
				} else {
					return path.Base(chunks[0])
				}
			})(),
			path: strings.Replace(chunks[0], "file://", "", 1),
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

func main() {
	setupTerminal()

	initialSearchText := ""

	if len(os.Args) > 1 {
		initialSearchText = os.Args[1]
	}

	isPrintShellScript := false
	isPrintShellEvalScript := false
	showVersion := false
	alias := "bcd"

	for _, arg := range os.Args[1:] {
		if arg == "--shell" {
			isPrintShellScript = true
		} else if arg == "--eval" {
			isPrintShellEvalScript = true
		} else if arg == "--version" {
			showVersion = true
		} else {
			alias = arg
		}
	}

	if showVersion {
		fmt.Println("bookmark-cd " + version)
		return
	}

	if isPrintShellScript {
		if isPrintShellEvalScript {
			fmt.Fprintf(os.Stdout, "\neval \"$(bookmark-cd --shell %s)\"\n", alias)
		} else {
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
		}

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
