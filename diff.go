package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type Hunk struct {
	Start int
	Count int
}

type Lines [][]byte

func (l Lines) String() string {
	return fmt.Sprintf("%s", bytes.Join(l, []byte{'\n'}))
}

type HunkPair struct {
	Removed Hunk
	Added   Hunk
	diff    Lines
}

type Hunks []*HunkPair

func (h Hunks) String() string {
	var lines []string
	for _, h := range h {
		lines = append(lines, fmt.Sprintf("%s", bytes.Join(h.diff, []byte{'\n'})))
	}
	return strings.Join(lines, "\n")
}

type Diff struct {
	// Total number of lines added
	Added int
	// Total number of lines removed
	Removed int
	Hunks   Hunks
}

func (d *Diff) String() string {
	return fmt.Sprintf("Added: %d\n"+
		"Removed: %d\n"+
		"Hunks: %s",
		d.Added, d.Removed, d.Hunks)
}

func NewDiff(r io.Reader) (Diff, error) {
	d := Diff{}
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return d, err
	}

	var currHunkPair *HunkPair
	buf = bytes.Replace(buf, []byte{'\r'}, nil, -1) // get rid of CR
	for _, line := range bytes.Split(buf, []byte{'\n'}) {
		if !bytes.HasPrefix(line, HUNK_PREFIX) {
			if len(line) > 0 && currHunkPair != nil {
				currHunkPair.diff = append(currHunkPair.diff, line)
			}
			continue
		}
		chunks := bytes.Split(line, SPACE)
		if len(chunks) < 4 {
			return d, fmt.Errorf("invalid line: %s", line)
		}
		removed := toHunk(chunks[1])
		added := toHunk(chunks[2])
		currHunkPair = &HunkPair{
			Removed: removed,
			Added:   added,
			diff:    [][]byte{line},
		}
		d.Hunks = append(d.Hunks, currHunkPair)
		d.Added += added.Count
		d.Removed += removed.Count
	}

	return d, nil
}

func hunkPair(rstart, rend, astart, aend int, lines string) *HunkPair {
	var diff [][]byte
	for _, line := range strings.Split(lines, "\n") {
		diff = append(diff, []byte(line))
	}
	return &HunkPair{
		Removed: Hunk{Start: rstart, Count: rend},
		Added:   Hunk{Start: astart, Count: aend},
		diff:    diff,
	}
}

// Reference: https://www.gnu.org/software/diffutils/manual/html_node/Detailed-Unified.html#Detailed-Unified
func toHunk(numbers []byte) Hunk {
	var start, count int

	lineRange := numbers[1:]
	if bytes.Contains(lineRange, COMMA) {
		// Many lines
		i := bytes.Index(lineRange, COMMA)
		start = asInt(lineRange[0:i])
		count = asInt(lineRange[i+1:])
	} else {
		// One line change
		start = asInt(lineRange)
		count = 1
	}
	return Hunk{
		Start: start,
		Count: count,
	}
}

func (d *Diff) LinesChanged() ([]int, []int) {
	var removed []int
	var added []int

	return removed, added
}
