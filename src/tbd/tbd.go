package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

/*
 tbd            - prints all current tasks
 tbd tags       - prints all current tags
 tbd #tagname   - prints all current tasks with #tagname
*/

var (
	// Can be used to extract tags from a line of text
	extract *regexp.Regexp
)

func init() {
	// (?:^|\\s)      - begin from the start or a white space, non-capturing
	// (#|@)          - capture the tag's type
	// ([^\\s]+?)     - capture the tag's value
	// (?:\\.|\\s|$)  - complete with a dot, a white space or the end, non-capturing
	extract = regexp.MustCompile("(?:^|\\s)(#|@)([^\\s]+?)(?:\\.|\\s|$)")
}

func main() {

	var (
		exitCode int
		handlers []handler = []handler{counting()}
	)

	// Exit handler
	defer func() {
		os.Exit(exitCode)
	}()

	// Parse command line elements and register handlers

	// Open default input
	input, err := os.Open("tbdata")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to open tbdata.", err.Error())
		exitCode = 1
		return
	}

	// Input close handler
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
		handle(handlers, &task{tags: parse(line), value: line})

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
	at   []string
}

// Parse the input for any tags
func parse(line string) (result tags) {
	for _, submatch := range extract.FindAllStringSubmatch(line, -1) {
		switch submatch[1] {
		case "#":
			result.hash = append(result.hash, submatch[2])
		case "@":
			result.at = append(result.at, submatch[2])
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

func (t tasks) Len() int {
	return len(t)
}

func (t tasks) Less(i, j int) bool {
	return t[i].nth < t[j].nth
}

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

// Returns a counting handler closure that sets the task's nth field
func counting() handler {
	var at int = 0
	return func(t *task) action {
		t.nth = at
		at++
		return action{}
	}
}

// Returns a grouping handler closure that stores tasks in tag chains
func grouping() handler {
	// Store dependency chains per tag type
	var chains = make(map[string]tasks)

	return func(t *task) action {
		for _, tag := range t.hash {
			// The previous tasks for current tag
			prev := chains[tag]

			// Point the current task to the last one with a matching tag, if any
			if l := len(prev); l > 0 {
				t.depends = append(t.depends, prev[l-1])
			}

			// Add the current task for the current tag
			chains[tag] = append(chains[tag], t)
		}
		return action{} // default, no action
	}
}

// Returns a matching handler closure that stores tasks in tag chains
func matching(tag string) handler {
	return func(t *task) action {
		for _, v := range t.hash {
			if v == tag {
				return action{}
			}
		}
		return action{stop: true}
	}
}
