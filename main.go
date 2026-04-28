package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func main() {
	var (
		pattern   string
		hasHeader bool
		sep       string
		invert    bool
	)

	flag.StringVar(&pattern, "pattern", "", "regexp applied on header fields (with -header) or on every field (without -header)")
	flag.BoolVar(&hasHeader, "header", false, "treat the first line as a header; pattern selects column names")
	flag.StringVar(&sep, "sep", ";", "field separator: any single character, 'space' (whitespace split), 'tab' or '\\t', 'auto' (whitespace)")
	flag.BoolVar(&invert, "invert", false, "invert selection: keep columns that do NOT match the pattern")
	flag.Parse()

	if pattern == "" {
		fmt.Fprintln(os.Stderr, "error: -pattern is required")
		flag.Usage()
		os.Exit(2)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid regexp %q: %v\n", pattern, err)
		os.Exit(2)
	}

	delim, mode, err := parseSeparator(sep)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid separator %q: %v\n", sep, err)
		os.Exit(2)
	}

	match := func(s string) bool {
		return re.MatchString(s) != invert
	}

	var runErr error
	if mode == "csv" {
		runErr = processCSV(os.Stdin, os.Stdout, match, hasHeader, delim)
	} else {
		runErr = processText(os.Stdin, os.Stdout, match, hasHeader)
	}

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", runErr)
		os.Exit(1)
	}
}

func parseSeparator(s string) (rune, string, error) {
	switch s {
	case "auto":
		return 0, "text", nil
	case "space":
		return ' ', "text", nil
	case `\t`, "tab":
		return '\t', "csv", nil
	default:
		runes := []rune(s)
		if len(runes) != 1 {
			return 0, "", fmt.Errorf("must be a single character, got %q", s)
		}
		return runes[0], "csv", nil
	}
}

// processCSV handles delimited files (CSV, TSV, PSV, …).
func processCSV(in io.Reader, out io.Writer, match func(string) bool, hasHeader bool, delim rune) error {
	r := csv.NewReader(in)
	r.Comma = delim
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	w := csv.NewWriter(out)
	w.Comma = delim
	defer w.Flush()

	var selected []int
	firstRow := true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if firstRow && hasHeader {
			selected = matchingIndexes(record, match)
			if err := w.Write(pickFields(record, selected)); err != nil {
				return err
			}
			firstRow = false
			continue
		}

		firstRow = false

		if hasHeader {
			if err := w.Write(pickFields(record, selected)); err != nil {
				return err
			}
		} else {
			if err := w.Write(filterFields(record, match)); err != nil {
				return err
			}
		}
	}

	w.Flush()
	return w.Error()
}

// processText handles whitespace-separated text (e.g. command output, /proc files).
func processText(in io.Reader, out io.Writer, match func(string) bool, hasHeader bool) error {
	scanner := bufio.NewScanner(in)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var selected []int
	firstLine := true

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			fmt.Fprintln(out)
			continue
		}

		if firstLine && hasHeader {
			selected = matchingIndexes(fields, match)
			fmt.Fprintln(out, strings.Join(pickFields(fields, selected), " "))
			firstLine = false
			continue
		}

		firstLine = false

		if hasHeader {
			fmt.Fprintln(out, strings.Join(pickFields(fields, selected), " "))
		} else {
			fmt.Fprintln(out, strings.Join(filterFields(fields, match), " "))
		}
	}

	return scanner.Err()
}

func matchingIndexes(fields []string, match func(string) bool) []int {
	idx := make([]int, 0, len(fields))
	for i, f := range fields {
		if match(f) {
			idx = append(idx, i)
		}
	}
	return idx
}

func pickFields(fields []string, indexes []int) []string {
	out := make([]string, 0, len(indexes))
	for _, i := range indexes {
		if i >= 0 && i < len(fields) {
			out = append(out, fields[i])
		}
	}
	return out
}

func filterFields(fields []string, match func(string) bool) []string {
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if match(f) {
			out = append(out, f)
		}
	}
	return out
}
