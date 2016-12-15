// hello
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) <= 1 {
		bail("Usage: git check-diff <file>")
	}

	checkDiff(os.Args[1])

}

type MergeBaseTags []string

func (m MergeBaseTags) Len() int           { return len(m) }
func (m MergeBaseTags) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MergeBaseTags) Less(i, j int) bool { return getTagNumber(m[i]) < getTagNumber(m[j]) }

func getTagNumber(mbtag string) int {
	if !strings.HasPrefix(mbtag, "MERGE_BASE_") {
		panic(fmt.Sprintf("%s is not a MERGE_BASE tag", mbtag))
	}
	chunks := strings.Split(mbtag, "_")
	if len(chunks) != 3 {
		panic(fmt.Sprintf("%s do not match MERGE_BASE_N pattern", mbtag))
	}

	n, err := strconv.Atoi(chunks[2])
	if err != nil {
		panic(fmt.Sprintf("%s: %v", mbtag, err))
	}
	return n
}

func checkDiff(file string) {
	var (
		HUNK_REMOVED = []byte{'@', '@', ' ', '-'}
		SPACE        = []byte{' '}
		COMMA        = []byte{','}
	)

	blame := getBlame(file)
	commitsAffected := map[string][]string{}

	for _, line := range linesFrom("git", "diff", "-U0", "--", file) {
		if !bytes.HasPrefix(line, HUNK_REMOVED) {
			continue
		}

		chunks := bytes.Split(line, SPACE)
		if len(chunks) <= 2 {
			bail("invalid line: %s", line)
		}

		lineRange := chunks[1][1:]
		if bytes.Contains(lineRange, COMMA) {
			i := bytes.Index(lineRange, COMMA)
			from := asInt(lineRange[0:i])
			to := asInt(lineRange[i+1:])
			if to == 0 {
				// fmt.Printf("   linerange %s\n", lineRange) // DEBUG
				// no lines removed, just new lines added
				continue
				// TODO show  merge base from the "from" line's
				// commit (need to deal with addition at top of
				// file, in which case `from` is 0
				// TODO how do we deal with addition at end of
				// file
			}

			for lnum := from; lnum < from+to-1; lnum++ {
				if lnum < len(blame) {
					commitsAffected[blame[lnum].sha1()] = nil
				} else {
					fmt.Printf("DEBUG out of bound len(blame) = %d, lnum %d\n", len(blame), lnum)
				}
			}
		} else {
			lnum := asInt(lineRange)
			commitsAffected[blame[lnum].sha1()] = nil

		}
	}
	fmt.Printf("Commits affected:\n")
	tagsSeen := map[string]int{}
	nCommits := len(commitsAffected)
	for sha1, _ := range commitsAffected {
		tags := findMergeBaseTags(sha1)
		commitsAffected[sha1] = tags
		for _, tag := range tags {
			tagsSeen[tag]++
		}
		fmt.Printf("\t%s\n", sha1)
	}

	var tags MergeBaseTags
	fmt.Printf("Common tag:\n")
	for tag, count := range tagsSeen {
		if count == nCommits {
			tags = append(tags, tag)
		}
	}

	sort.Sort(tags)
	for i, tag := range tags {
		fmt.Printf("\t%s\n", tag)
		_ = i
		if i >= 10 && i < len(tags)-1 {
			fmt.Printf("\t... %d more\n", len(tags)-(i+1))
			break
		}
	}
}

func findMergeBaseTags(sha1 string) []string {
	var tags []string
	for _, line := range linesFrom("git", "tag", "--contains", sha1, "-l", "MERGE_BASE_*") {
		if len(line) > 0 {
			tags = append(tags, string(line))
		}
	}
	return tags
}

func asInt(buf []byte) int {
	n, err := strconv.Atoi(string(buf))
	if err != nil {
		bail("%s: %v", buf, err)
	}
	return n
}

type LineBlame []byte

func (lb LineBlame) sha1() string {
	i := bytes.Index(lb, []byte{' '})
	return string(lb[0:i])
}

type Blame []LineBlame

func getBlame(file string) Blame {
	blame := Blame{[]byte("NIL")}
	for _, line := range linesFrom("git", "blame", "-l", "--root", "-r", "HEAD", file) {
		lblame := LineBlame(line)
		blame = append(blame, lblame)
	}
	return blame
}

func linesFrom(command string, arg ...string) [][]byte {
	return bytes.Split(run(command, arg...), []byte{'\n'})
}

func run(name string, arg ...string) []byte {
	buf, err := exec.Command(name, arg...).Output()
	if err != nil {
		bail("%v", err)
	}
	return buf
}

func bail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
