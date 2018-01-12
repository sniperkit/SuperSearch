package gitignore

import (
	"bufio"
	"os"
	"strings"
)

var (
	commentPrefix = "#"
	includePrefix = "!"
	zeroToManyDir = "**"
)

// Maches paths based on defined include/exclude patterns
type Matcher struct {
	includes *PatternMatcher
	excludes *PatternMatcher
}

// Define gitignore match results
type MatchResult int

const (
	Include, Exclude NoMatch = iota
)

// Returns true if the pattern
func (m *Matcher) Match(path string) bool {
	return false
}

// Read a file into a Matcher instance
func ReadPatterns(path string) Matcher {
	f, err := os.Open(path)
	var m Matcher
	if err != nil {
		return m
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line.TrimSpace()) == 0 || strings.HasPrefix(line, commentPrefix) {
			continue
		}
		if strings.HasPrefix(line, includePrefix) {
			m.includes = m.includes.AddPattern(line[1:])
		} else {
			m.excludes = m.excludes.AddPattern(line)
		}
		m.parsePattern(line)
	}
	return m
}
