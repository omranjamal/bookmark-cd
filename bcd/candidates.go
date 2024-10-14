package bcd

import (
	"bufio"
	"fmt"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
)

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
