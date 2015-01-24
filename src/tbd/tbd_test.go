package main

import (
	"testing"
)

func TestParse(t *testing.T) {
	input := "This is a #sample task, @done."
	tags := parse(input)
	if len(tags.hash) != 1 || tags.hash[0] != "sample" {
		t.Error("Unexpect tags", tags)
	}
	if len(tags.at) != 1 || tags.at[0] != "done" {
		t.Error("Unexpect tags", tags)
	}
}
