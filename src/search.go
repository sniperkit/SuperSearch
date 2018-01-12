package search

import (
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/exp/mmap"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
	// "github.com/wellsjo/search/src/ignore2"
	// "github.com/monochromegane/go-gitignore"
	// "gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
)

var (
	ignoreFilePatterns   = []string{}
	globalIgnorePatterns = []*regexp.Regexp{}
	concurrency          = 64
	highlightMatch       = color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold)
	highlightFile        = color.New(color.FgCyan).Add(color.Bold)
	highlightNumber      = color.New(color.FgGreen).Add(color.Bold)
)

type PrintData struct {
	file string
	data string
}

type SuperSearch struct {
	searchRegexp *regexp.Regexp
	location     string
	print        chan *PrintData
	finished     chan string
	done         chan bool
	sem          chan bool
	wg           *sync.WaitGroup
}

func NewSuperSearch() {
	opts := GetOptions()
	Debug("Searching", opts.location, "for", opts.pattern)
	Debug("Concurrency", *opts.concurrency)
	ss := &SuperSearch{
		searchRegexp: regexp.MustCompile(opts.pattern),
		location:     opts.location,
		print:        make(chan *PrintData),
		finished:     make(chan string),
		done:         make(chan bool),
		sem:          make(chan bool, *opts.concurrency),
		wg:           new(sync.WaitGroup),
	}
	go ss.runPrinter()
	ss.run()
	ss.wg.Wait()
	ss.wg.Add(1)
	ss.done <- true
	ss.wg.Wait()
}

func (ss *SuperSearch) run() {
	fi, err := os.Stat(ss.location)
	if err != nil {
		fmt.Println(err)
		return
	}
	ss.wg.Add(1)
	switch mode := fi.Mode(); {
	case mode.IsDir():
		ss.ScanDir(ss.location)
	case mode.IsRegular():
		ss.sem <- true
		ss.SearchFile(ss.location)
	}
}

func (ss *SuperSearch) runPrinter() {
	var dataToPrint = make(map[string][]string)
	var finishedFiles = make(map[string]bool)
	var canFinish bool = false
	var curFile string
OUTER:
	for {
		select {
		case s := <-ss.print:
			dataToPrint[s.file] = append(dataToPrint[s.file], s.data)
		case finished := <-ss.finished:
			finishedFiles[finished] = true
		case <-ss.done:
			canFinish = true
		default:
			if len(dataToPrint[curFile]) > 0 {
				fmt.Print(strings.Join(dataToPrint[curFile], ""))
				delete(dataToPrint, curFile)
			}
			if finishedFiles[curFile] {
				delete(finishedFiles, curFile)
				fmt.Println()
				curFile = ""
			} else if canFinish && len(dataToPrint) == 0 {
				break OUTER
			}
			if curFile == "" {
				for i := range dataToPrint {
					curFile = i
					highlightFile.Println(curFile)
					break
				}
			}
		}
	}
	ss.wg.Done()
}

func (ss *SuperSearch) ScanDir(dir string) {
	Debug("Scanning directory", dir)
	dirInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	// m := ignore.ReadPatterns(filepath.Join(dir, ".gitignore"))

	// DIRLOOP:
	for _, fi := range dirInfo {
		if fi.Name()[0] == '.' {
			continue
		}
		path := filepath.Join(dir, fi.Name())
		if fi.IsDir() {
			// if m != nil {
			// 	x := strings.Split(path, string(filepath.Separator))
			// 	Debug("Testing", x)
			// 	r := m.Match(x, true)
			// 	Debug("Matching", path, r)
			// }
			// for _, m := range matchers {
			// 	if m.Match(path, true) {
			// 		Debug("ignoring directory", path)
			// 		continue DIRLOOP
			// 	}
			// }
			ss.wg.Add(1)
			go ss.ScanDir(path)
		} else if fi.Mode().IsRegular() {
			// if m != nil {
			// 	x := strings.Split(path, string(filepath.Separator))
			// 	Debug("Testing", x)
			// 	r := m.Match(x, false)
			// 	Debug("Matching", path, r)
			// }
			// for _, m := range matchers {
			// if m.Match(path, false) {
			// 	Debug("ignoring file", path)
			// 	continue DIRLOOP
			// }
			// }
			ss.wg.Add(1)
			go func() {
				ss.sem <- true
				ss.SearchFile(path)
			}()
		}
	}
	Debug("Goroutine ScanDir", dir, "finished")
	ss.wg.Done()
}

func (ss *SuperSearch) SearchFile(path string) {
	Debug("Goroutine created. Searching file", path)
	file, err := mmap.Open(path)
	if err != nil {
		Debug("Failed to open file with mmap", path)
		panic(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
		Debug("Closing file search goroutine", path)
		ss.finished <- path
		<-ss.sem
		ss.wg.Done()
	}()
	if isBin(file) {
		Debug("Skipping binary file", path)
		return
	}
	if file.Len() == 0 {
		Debug("Skipping empty file", path)
		return
	}
	lastIndex := 0
	lineNo := 1
	buf := make([]byte, file.Len())
	bytesRead, err := file.ReadAt(buf, 0)
	if err != nil {
		Debug("Failed to read file", path+".", "Read", bytesRead, "bytes.")
		panic(err)
	}
	for i := 0; i < len(buf); i++ {
		if buf[i] == '\n' {
			var line = buf[lastIndex:i]
			ixs := ss.searchRegexp.FindAllIndex(line, -1)
			var output string
			if ixs != nil {
				output = highlightNumber.Sprint(lineNo, ":")
				lastIndex := 0
				for _, i := range ixs {
					output += fmt.Sprint(string(line[lastIndex:i[0]]))
					output += highlightMatch.Sprint(string(line[i[0]:i[1]]))
					lastIndex = i[1]
				}
				output += fmt.Sprintln(string(line[lastIndex:]))
			}
			if len(output) > 0 {
				p := PrintData{
					file: path,
					data: output,
				}
				ss.print <- &p
			}
			lastIndex = i + 1
			lineNo++
		}
	}
}

// TODO come up with a better way to do this
func isBin(file *mmap.ReaderAt) bool {
	var offsetLen int64 = int64(file.Len()) / 4
	var offset int64 = 0
	var buf = make([]byte, 4)
	for i := 0; i < 4; i++ {
		file.ReadAt(buf, offset)
		if !utf8.Valid(buf) {
			return true
		}
		offset += offsetLen
	}
	return false
}
