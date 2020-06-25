package csvxlib

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Writer interface {
	Write([]string) error
	Flush()
}

type TableWriter struct{ *tabwriter.Writer }

func NewTableWriter(w io.Writer) *TableWriter {
	fmt.Fprintln(w, "yeah")
	return &TableWriter{Writer: tabwriter.NewWriter(w, 5, 4, 2, ' ', tabwriter.TabIndent)}
}

func (t *TableWriter) Write(line []string) error {
	fmt.Fprintln(t.Writer, strings.Join(line, "\t"))
	return nil
}

func (t *TableWriter) Flush() {
	if err := t.Writer.Flush(); err != nil {
		fmt.Println("Gah")
	}
}

type RawCSVWriter struct{ io.Writer }

func (r RawCSVWriter) Write(line []string) error {
	fmt.Fprintln(r.Writer, strings.Join(line, ","))
	return nil
}
func (RawCSVWriter) Flush() {}

type Transformer interface {
	Transform([]string) ([]string, error)
}

type IncludeIndicies []int

func (i IncludeIndicies) Transform(in []string) ([]string, error) {
	if len(i) == 0 {
		return in, nil
	}
	out := make([]string, 0, len(in))
	for _, j := range i {
		if j >= len(in) {
			return nil, fmt.Errorf("index %d not found", j)
		}
		out = append(out, in[j])
	}
	return out, nil
}

func Pipe(r io.Reader, out Writer, transforms ...Transformer) error {
	csvr := csv.NewReader(r)
	csvr.FieldsPerRecord = -1
	for {
		record, err := csvr.Read()
		if err != nil {
			return err
		}
		for _, t := range transforms {
			if record, err = t.Transform(record); err != nil {
				return fmt.Errorf("problem with record=%#v err=%w", record, err)
			}
		}
		out.Write(record)
	}
}
