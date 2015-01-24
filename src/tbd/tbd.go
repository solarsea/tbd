// tbd, a #tag-based dependencies filter useful for low ceremony task management
// see README.md for usage tips
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
	extract = regexp.MustCompile("(#)([^\\s]+?)(?:[,\\.\\s]|$)")
}

func main() {

	var (
		exitCode int
		handlers []handler = []handler{counting(), tracing()}
	)

	// Exit handler
	defer func() {
		os.Exit(exitCode)
	}()

	// Create the final handler and stage the output
	collect, output := collecting()
	defer func() {
		seq := make(tasks, 0, len(output))
		for k, _ := range output {
			seq = append(seq, k)
		}
		sort.Stable(seq)
		for _, v := range seq {
			fmt.Println(v)
		}
	}()

	// Parse command line elements and register handlers
	if argc := len(os.Args); argc == 1 {
		handlers = append(handlers, cutoff())
	} else {
		var tags = make([]string, argc-1)
		for _, arg := range os.Args[1:] {
			tags = append(tags, arg)
		}
		handlers = append(handlers, matching(tags))
	}

	// Register the final handler
	handlers = append(handlers, collect)

	// Open default input
	input, err := os.Open("tbdata")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to open tbdata.", err.Error())
		exitCode = 1
		return
	}

	// Make sure to close the input
	defer func() {
		err := input.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to close input.", err.Error())
			exitCode = 1
		}
	}()

	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')

		line = strings.TrimSpace(line)
		if len(line) > 0 {
			handle(handlers, &task{tags: parse(line), value: line})
		}

		// Process errors after handling as ReadString can return both data and error f
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintln(os.Stderr, "Error reading input.", err.Error())
			exitCode = 1
		}
	}
}

// The tags struct contains hash (#) and at (@) tags
type tags struct {
	hash []string
}

// Parse the input for any tags
func parse(line string) (result tags) {
	for _, submatch := range extract.FindAllStringSubmatch(line, -1) {
		switch submatch[1] {
		// Other tag types can be added as well
		case "#":
			result.hash = append(result.hash, submatch[2])
		}
	}
	return
}

// The task struct contains a task description and its tags
type task struct {
	tags
	nth     int
	value   string
	depends tasks
}

// Alias for a slice of task pointer
type tasks []*task

// as per sort.Interface
func (t tasks) Len() int {
	return len(t)
}

// as per sort.Interface
func (t tasks) Less(i, j int) bool {
	return t[i].nth < t[j].nth
}

// as per sort.Interface
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

// Execute handlers against a task
func handle(handlers []handler, t *task) {
	for _, h := range handlers {
		act := h(t)
		if act.stop {
			return
		}
	}
}

// Returns a counting handler closure that sets the tasks' nth field
func counting() handler {
	var at int = 0
	return func(t *task) action {
		t.nth = at
		at++
		return action{}
	}
}

// Returns a tracing handler closure that sets the tasks' depends field
func tracing() handler {
	// Store the last task per tag type
	var last = make(map[string]*task)
	return func(t *task) action {
		for _, tag := range t.hash {
			if prev, ok := last[tag]; ok {
				t.depends = append(t.depends, prev)
			}
			last[tag] = t
		}
		return action{} // default, no action
	}
}

// Returns handler closure that rejects tasks with already-seen tags
func cutoff() handler {
	var seen = make(map[string]bool)
	return func(t *task) (result action) {
		for _, tag := range t.hash {
			if seen[tag] {
				result.stop = true
			} else {
				seen[tag] = true
			}
		}
		return
	}
}

// Returns a matching handler closure that filters tasks not matching atleast one tag
func matching(tags []string) handler {
	var allowed = make(map[string]bool)
	for _, tag := range tags {
		allowed[tag] = true
	}

	return func(t *task) action {
		for _, tag := range t.hash {
			if allowed[tag] {
				return action{}
			}
		}
		return action{stop: true}
	}
}

// Returns a handler that stores every seen task in the map
func collecting() (handler, map[*task]struct{}) {
	var seen = make(map[*task]struct{})
	return func(t *task) action {
		seen[t] = struct{}{}
		return action{}
	}, seen
}
