package gitignore

func Test() {
	p := NewPatternTrie()
	p.AddPattern("hello")
	p.AddPattern("world")
	p.AddPattern("heck")
}
