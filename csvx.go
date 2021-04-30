package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
		output           = flag.String("output", "csv", "Output ["+strings.Join(outputs.Keys(), ",")+"]")
		strict           = flag.Bool("strict", true, "Turn on strict mode")
		printHeaders     = flag.Bool("print-headers", false, "Turn on Column Numbers")
	)
	flag.Var(&include, "i", "indicies to include")
	flag.Var(&exclude, "x", "indicies to exclude")
	flag.Var(&delim, "d", "delimiter")
	flag.Parse()
	//
	// Set up the CSV Reader
	r := newInputReader()
	defer r.Close()
	csvr := csv.NewReader(bufio.NewReader(r))
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
	// Tracker
	var tracker csvxlib.Tracker
	if *printHeaders {
		transforms = append(transforms, &tracker)
	}
	var (
		w   = newWriter(out)
		err = csvxlib.Pipe(csvr, w, transforms...)
	)
	w.Flush()
	if err != nil && err != io.EOF {
		log.Fatal("Problem processing csv", err)
	}
	if *printHeaders {
		headerOut := csvxlib.NewTableWriter(out)
		for i, h := range tracker.Headers {
			headerOut.Write([]string{strconv.Itoa(i), h})
		}
		headerOut.Flush()
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

func newInputReader() io.ReadCloser {
	var (
		args    = flag.Args()
		stat, _ = os.Stdin.Stat()
	)
	// We got something from stdin
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if len(args) > 0 {
			log.Fatal("no arguments expected when reading from stdin")
		}
		return ioutil.NopCloser(os.Stdin)
	}
	// Let's open a file
	if len(args) == 1 {
		r, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("could not open file %q err=%q", args[0], err)
		}
		return r
	}
	log.Fatal("expected 1 argument of the csv file to process or to read from stdin")
	return nil
}
