package csvxlib

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Writer interface {
	Write([]string) error
	Flush()
}

type RecordReader interface {
	Read() ([]string, error)
}

type TableWriter struct{ tabwriter.Writer }

func NewTableWriter(w io.Writer) *TableWriter {
	var tw TableWriter
	tw.Init(w, 5, 4, 2, ' ', tabwriter.TabIndent)
	return &tw
}

func (t *TableWriter) Flush() { t.Writer.Flush() }
func (t *TableWriter) Write(line []string) error {
	fmt.Fprintln(&t.Writer, strings.Join(line, "\t"))
	return nil
}

type RawCSVWriter struct{ io.Writer }

func (RawCSVWriter) Flush() {}
func (r RawCSVWriter) Write(line []string) error {
	fmt.Fprintln(r.Writer, strings.Join(line, ","))
	return nil
}

type Transformer interface {
	Transform([]string) ([]string, error)
}

type ExcludeIndicies struct {
	List []int
	set  map[int]struct{}
}

func (i *ExcludeIndicies) Transform(in []string) ([]string, error) {
	if len(i.List) == 0 {
		return in, nil
	}
	if i.set == nil {
		i.set = map[int]struct{}{}
		for _, j := range i.List {
			i.set[j] = struct{}{}
		}
	}
	out := make([]string, 0, len(in))
	for j, s := range in {
		if _, ok := i.set[j]; !ok {
			out = append(out, s)
		}
	}
	return out, nil
}

type IncludeIndicies struct {
	List   []int
	Strict bool
}

func (i IncludeIndicies) Transform(in []string) ([]string, error) {
	if len(i.List) == 0 {
		return in, nil
	}
	out := make([]string, 0, len(in))
	for _, j := range i.List {
		if j >= len(in) {
			if i.Strict {
				return nil, fmt.Errorf("index %d not found", j)
			}
			continue
		}
		out = append(out, in[j])
	}
	return out, nil
}

func Pipe(r RecordReader, out Writer, transforms ...Transformer) error {
	for {
		record, err := r.Read()
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
