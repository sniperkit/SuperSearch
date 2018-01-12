package gitignore

// Create new PatternMatcher and return pointer to it
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		head: make(map[string]*patternNode),
	}
}

// Defines a set of patterns, stored in a trie data structure for fastest
// insert/lookup
type PatternMatcher struct {
	head map[string]*patternNode
}

// Adds a pattern to the pattern matcher (trie insert)
func (p *PatternMatcher) AddPattern(p string) {
	var curNode = p.head
	for i := 0; i < len(p); i++ {
		isEnd := i == len(p)-1
		if next, exists := curNode[p[i]]; exists {
			if isEnd {
				next.end = true
				return
			} else {
				curNode = n
				continue
			}
		} else {
			curNode[p[i]] = patternNode{
				end:  isEnd,
				next: make(map[string]patternNode),
			}
			if !isEnd {
				curNode = curNode[p[i]]
			}
		}
	}
}

// Represents an individual character pattern character in a pattern matcher
type patternNode struct {
	end  bool
	next map[string]*patternNode
}

// Add a pattern node to PatternMatcher, and return a pointer to the new node
func (p *patternNode) addNode(val string, end bool) *patternNode {
	p.next[val] = patternNode{
		end,
		next: make(map[string]patternNode),
	}
	return &p.next[val]
}
