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
		handlers []handler = make([]handler, 1)
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
		handle(handlers, task{parse(line), line})

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
func parse(line string) tags {
	var result tags

	for _, submatch := range extract.FindAllStringSubmatch(line, -1) {
		switch submatch[1] {
		case "#":
			result.hash = append(result.hash, submatch[2])
		case "@":
			result.at = append(result.at, submatch[2])
		}
	}

	return result
}

// The task struct contains a task description and its tags
type task struct {
	tags
	value string
}

// The action struct contains flags that determine what to do with a task
type action struct {
	// Stop further processing of the task
	stop bool
}

// An alias interface for all task handler functions
type handler func(task) action

// Execute handlers against a task
func handle(handlers []handler, t task) {
	for _, h := range handlers {
		act := h(t)
		if act.stop {
			return
		}
	}
}
