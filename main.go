package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

func main() {
	var (
		indicies filterIndicies
		output             = flag.String("output", "", "Output [raw,table]")
		br                 = bufio.NewReader(os.Stdin)
		csvr               = csv.NewReader(br)
		out      io.Writer = os.Stdout
		csvw     csvwriter
	)
	defer func() { csvw.Flush() }()
	csvr.FieldsPerRecord = -1
	flag.Var(&indicies, "i", "indicies to remove")
	flag.Parse()
	switch *output {
	default:
		log.Fatalf("Invalid output format=%q", *output)
	case "":
		csvw = csv.NewWriter(out)
	case "raw":
		csvw = rawCSVWriter{Writer: out}
	case "tab", "table":
		csvw = newTableWriter(out)
	}
	for {
		record, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if record, err = indicies.Apply(record); err != nil {
			log.Fatal(err)
		}
		csvw.Write(record)
	}
}

type tableWriter struct{ *tabwriter.Writer }

func newTableWriter(w io.Writer) tableWriter {
	return tableWriter{Writer: tabwriter.NewWriter(w, 5, 4, 2, ' ', tabwriter.TabIndent)}
}
func (t tableWriter) Write(line []string) error {
	fmt.Fprintln(t.Writer, strings.Join(line, "\t"))
	return nil
}
func (t tableWriter) Flush() { t.Writer.Flush() }

type rawCSVWriter struct{ io.Writer }

func (r rawCSVWriter) Write(line []string) error {
	fmt.Fprintln(r.Writer, strings.Join(line, ","))
	return nil
}
func (r rawCSVWriter) Flush() {}

type csvwriter interface {
	Write([]string) error
	Flush()
}

type filterIndicies []int

func (f *filterIndicies) String() string {
	return fmt.Sprintf("%#v", *f)
}
func (f *filterIndicies) Apply(in []string) ([]string, error) {
	if f == nil || len(*f) == 0 {
		return in, nil
	}
	out := make([]string, len(*f))
	for i, j := range *f {
		if j < len(in) {
			out[i] = in[j]
		} else {
			return nil, fmt.Errorf("index %d not in row=%#v", j, in)
		}
	}
	return out, nil
}
func (f *filterIndicies) Set(in string) error {
	for _, s := range strings.Split(in, ",") {
		i, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("could not parse %q as a number %w", s, err)
		}
		*f = append(*f, i)
	}
	return nil
}
