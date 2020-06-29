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
		out              io.Writer = os.Stdout
		include, exclude FlagIntSlice
		delim            FlagRune = ','
		transforms       []csvxlib.Transformer
		csvr             = csv.NewReader(bufio.NewReader(os.Stdin))
		output           = flag.String("output", "csv", "Output ["+strings.Join(outputs.Keys(), ",")+"]")
		strict           = flag.Bool("strict", true, "Turn on strict mode")
	)
	flag.Var(&include, "i", "indicies to include")
	flag.Var(&exclude, "x", "indicies to exclude")
	flag.Var(&delim, "d", "delimiter")
	flag.Parse()
	// Set up the CSV Reader
	csvr.FieldsPerRecord = -1
	csvr.Comma = rune(delim)
	// Grab the specified output format
	newWriter, ok := outputs[*output]
	if !ok {
		log.Fatalf("Invalid output format=%q must be one of [%s]", *output, strings.Join(outputs.Keys(), ","))
	}
	if len(include) > 0 {
		transforms = append(transforms, csvxlib.IncludeIndicies{
			List:   include,
			Strict: *strict,
		})
	}
	if len(exclude) > 0 {
		transforms = append(transforms, &csvxlib.ExcludeIndicies{List: exclude})
	}
	var (
		w   = newWriter(out)
		err = csvxlib.Pipe(csvr, w, transforms...)
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

type FlagRune rune

func (f *FlagRune) String() string { return fmt.Sprintf("%#v", *f) }
func (f *FlagRune) Set(in string) error {
	runes := []rune(in)
	if len(runes) != 1 {
		return fmt.Errorf("specify a single rune got=%q", in)
	}
	*f = FlagRune(runes[0])
	return nil
}

type OutputMap map[string]func(io.Writer) csvxlib.Writer

func (o OutputMap) Keys() (keys []string) {
	for key := range outputs {
		keys = append(keys, key)
	}
	return
}
