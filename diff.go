package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

type Hunk struct {
	Start int
	Count int
}

type HunkPair struct {
	Removed Hunk
	Added   Hunk
}
type Diff struct {
	// Total number of lines added
	Added int
	// Total number of lines removed
	Removed int
	Hunks   []HunkPair
}

func NewDiff(r io.Reader) (Diff, error) {
	d := Diff{}
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return d, err
	}

	buf = bytes.Replace(buf, []byte{'\r'}, nil, -1) // get rid of CR
	for _, line := range bytes.Split(buf, []byte{'\n'}) {
		if !bytes.HasPrefix(line, HUNK_PREFIX) {
			continue
		}
		chunks := bytes.Split(line, SPACE)
		if len(chunks) < 4 {
			return d, fmt.Errorf("invalid line: %s", line)
		}
		removed := toHunk(chunks[1])
		added := toHunk(chunks[2])
		d.Hunks = append(d.Hunks, HunkPair{
			Removed: removed,
			Added:   added,
		})
		d.Added += added.Count
		d.Removed += removed.Count
	}

	return d, nil
}

func hunkPair(rstart, rend, astart, aend int) HunkPair {
	return HunkPair{
		Removed: Hunk{Start: rstart, Count: rend},
		Added:   Hunk{Start: astart, Count: aend},
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
