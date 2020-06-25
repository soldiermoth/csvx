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

	"github.com/soldiermoth/csvx/csvxlib"
)

var (
	outputs = OutputMap{
		"csv":   func(w io.Writer) csvxlib.Writer { return csv.NewWriter(w) },
		"raw":   func(w io.Writer) csvxlib.Writer { return csvxlib.RawCSVWriter{Writer: w} },
		"table": func(w io.Writer) csvxlib.Writer { return csvxlib.NewTableWriter(w) },
	}
)

func main() {
	var (
		out        io.Writer = os.Stdout
		include    FlagIntSlice
		transforms []csvxlib.Transformer
		output     = flag.String("output", "csv", "Output ["+strings.Join(outputs.Keys(), ",")+"]")
	)
	flag.Var(&include, "i", "indicies to include")
	flag.Parse()
	newWriter, ok := outputs[*output]
	if !ok {
		log.Fatalf("Invalid output format=%q must be one of [%s]", *output, strings.Join(outputs.Keys(), ","))
	}
	if len(include) > 0 {
		transforms = append(transforms, csvxlib.IncludeIndicies(include))
	}
	var (
		br  = bufio.NewReader(os.Stdin)
		w   = newWriter(out)
		err = csvxlib.Pipe(br, w, transforms...)
	)
	defer w.Flush()
	if err != nil && err != io.EOF {
		log.Fatal("Problem processing csv", err)
	}
}

type FlagIntSlice []int

func (f *FlagIntSlice) String() string { return fmt.Sprintf("%#v", *f) }
func (f *FlagIntSlice) Set(in string) error {
	for _, s := range strings.Split(in, ",") {
		i, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("could not parse %q as a number %w", s, err)
		}
		*f = append(*f, i)
	}
	return nil
}

type OutputMap map[string]func(io.Writer) csvxlib.Writer

func (o OutputMap) Keys() (keys []string) {
	for key := range outputs {
		keys = append(keys, key)
	}
	return
}
