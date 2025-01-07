package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	_ "strings"
)

var depth int

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

type Entry interface {
	GetName() string
	IsLast() bool
	HasFinalParent() bool
}

type EntryDir struct {
	name           string
	isLast         bool
	hasFinalParent bool
}

func (e EntryDir) GetName() string {
	return e.name
}

func (e EntryDir) IsLast() bool {
	return e.isLast
}

func (e EntryDir) HasFinalParent() bool {
	return e.hasFinalParent
}

type EntryFile struct {
	name           string
	isLast         bool
	hasFinalParent bool
}

func (e EntryFile) GetName() string {
	return e.name
}

func (e EntryFile) IsLast() bool {
	return e.isLast
}

func (e EntryFile) HasFinalParent() bool {
	return e.hasFinalParent
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	depth++

	r, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open path: %s", err.Error())
	}

	entries, err := r.ReadDir(0)
	if err != nil {
		return fmt.Errorf("failed to read dir: %s", err.Error())
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var hasFinalParent bool
	for i := 0; i < len(entries); i++ {
		e := entries[i]
		if e.IsDir() {
			isLastEntry := false
			if i == depth-1 {
				isLastEntry = true
				if depth == 1 {
					hasFinalParent = true
				}
			}
			displayEntry := EntryDir{
				name: e.Name(),
				isLast: isLastEntry,
				hasFinalParent: hasFinalParent,
			}
			printEntry(out, displayEntry, depth)
			// fmt.Println(displayEntry.GetName())

			relPath := filepath.Join(path, e.Name())
			dirTree(out, relPath, printFiles)
		} else {
			// print files
		}
	}

	depth--

	return nil
}

func printEntry(out io.Writer, e Entry, depth int) {
	prefix := "├───"
	if e.IsLast() {
		prefix = "└───"
	}

	if depth > 1 {
		if !e.HasFinalParent() {
			fmt.Fprint(out, "│")
		}
		for i := 1; i < depth; i++ {
			fmt.Fprint(out, "\t")
		}
	}

	fmt.Fprintf(out, "%v%v\n", prefix, e.GetName())
}
