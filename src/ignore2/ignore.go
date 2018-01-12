package ignore

import (
	"bufio"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
)

var commentPrefix = "#"

// Read a file into a Matcher instance
func ReadPatterns(path string) gitignore.Matcher {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var ps = []gitignore.Pattern{}
	for scanner.Scan() {
		pattern := scanner.Text()
		if len(strings.TrimSpace(pattern)) == 0 || strings.HasPrefix(pattern, commentPrefix) {
			continue
		}
		ps = append(ps, gitignore.ParsePattern(pattern, nil))
	}
	return gitignore.NewMatcher(ps)
}
