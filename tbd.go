// tbd, a #tag-based dependency tool for task tracking
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	// Can be used to extract tags from a line of text
	extract *regexp.Regexp
)

func init() {
	// (#)([^\\s]+?)   - capture the tag's type and value
	// (?:\\.|\\s|$)   - complete the match, non-capturing
	extract = regexp.MustCompile("(##|#)([^\\s]+?)(?:[,\\.\\s]|$)")
}

func main() {

	var (
		exitCode int
		handlers = handlerChain{counting(), dependencies()}
	)

	// Parse command line elements and register handlers
	if len(os.Args) == 1 {
		handlers = append(handlers, cutoff())
	} else {
		handlers = append(handlers, matching(os.Args[1:]))
	}

	collect, output := collecting()
	handlers = append(handlers, collect)

	// Open default input
	input, err := os.Open("tbdata")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to open tbdata.", err.Error())
		os.Exit(1)
	}

	exitCode = process(input, handlers)

	for _, v := range output() {
		fmt.Println(v)
	}

	if err := input.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to close input.", err.Error())
		exitCode = 1
	}

	os.Exit(exitCode)
}

// Processes input through the provided handlers
func process(input io.Reader, handlers handlerChain) int {
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')

		line = strings.TrimSpace(line)
		if len(line) > 0 {
			handlers.do(&task{tags: parse(line), value: line, depends: make(tasks, 0, 8)})
		}

		// Process errors after handling as ReadString can return both data and error f
		if err != nil {
			if err == io.EOF {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Error reading input.", err.Error())
			return 1
		}
	}
}

// The tags struct contains hash (#) tags
type tags struct {
	provided  map[string]struct{}
	dependent map[string]struct{}
}

// Parse the input for any tags
func parse(line string) (result tags) {
	result.provided = make(map[string]struct{})
	result.dependent = make(map[string]struct{})

	//       | provides | depends
	// ------+----------+---------
	//     # |   #, ##  |  #
	//    ## |   #      |  ##

	for _, submatch := range extract.FindAllStringSubmatch(line, -1) {
		tag := submatch[2]
		switch submatch[1] {
		// Other tag types can be added as well
		case "#":
			result.provided[tag] = struct{}{}
			result.provided["#"+tag] = struct{}{}

			result.dependent[tag] = struct{}{}
		case "##":
			result.provided[tag] = struct{}{}

			result.dependent["#"+tag] = struct{}{}
		}
	}
	return
}

// The task struct contains a task description and its tags
type task struct {
	tags    tags
	nth     int
	value   string
	depends tasks
}

// per fmt.Stringer
func (t *task) String() string {
	return fmt.Sprintf("%d) %s", t.nth, t.value)
}

// Alias for a slice of task pointer
type tasks []*task

// per sort.Interface
func (t tasks) Len() int {
	return len(t)
}

// per sort.Interface
func (t tasks) Less(i, j int) bool {
	return t[i].nth < t[j].nth
}

// per sort.Interface
func (t tasks) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// The action struct contains flags that determine what to do with a task
type action struct {
	// Stop further processing of the task
	stop bool
}

// An alias interface for all task handler functions
type handler func(*task) action

// A slice of handlers
type handlerChain []handler

// Execute handlers against a task
func (handlers handlerChain) do(t *task) {
	for _, current := range handlers {
		act := current(t)
		if act.stop {
			return
		}
	}
}

// Returns a handler that sets the tasks' nth field
func counting() handler {
	var at int = 1
	return func(t *task) action {
		t.nth = at
		at++
		return action{}
	}
}

// Returns a handler that sets the tasks' depends field
func dependencies() handler {
	// Store the last task per tag
	var last = make(map[string]*task)

	return func(t *task) action {
		for tag, _ := range t.tags.dependent {
			if previousTask, ok := last[tag]; ok {
				t.depends = append(t.depends, previousTask)
			}
		}
		for tag, _ := range t.tags.provided {
			last[tag] = t
		}
		return action{} // default, no action
	}
}

// Returns a handler that rejects tasks with dependencies
func cutoff() handler {
	return func(t *task) action {
		if len(t.depends) > 0 {
			return action{stop: true}
		}
		return action{}
	}
}

// Returns a handler  that filters tasks not matching atleast one tag
func matching(tags []string) handler {
	var allowed = make(map[string]bool)
	for _, tag := range tags {
		allowed[tag] = true
	}

	return func(t *task) action {
		for tag, _ := range t.tags.provided {
			if allowed[tag] {
				return action{}
			}
		}
		return action{stop: true}
	}
}

// Returns a handler that stores every seen task and its dependencies
//    along with an ancillary function to retrieve those in a sorted order
func collecting() (handler, func() tasks) {
	var seen = make(map[*task]bool)

	collector := func(t *task) action {
		depth := tasks{t}
		for i := 0; i < len(depth); i++ {
			if !seen[depth[i]] {
				seen[depth[i]] = true
				depth = append(depth, depth[i].depends...)
			}
		}
		return action{}
	}

	retriever := func() tasks {
		seq := make(tasks, 0, len(seen))
		for k, _ := range seen {
			seq = append(seq, k)
		}
		sort.Stable(seq)
		return seq
	}

	return collector, retriever
}
