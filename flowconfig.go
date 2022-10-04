package youtubetoolkit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type FlowOption func(*flowconfig)

type flowconfig struct {
	stringSource func(errors chan<- error) <-chan string
	itemSink     func(errors chan<- error, input <-chan Item)
}

// SingleStringSource sets the source to only emit the param input string.
func SingleStringSource(input string) FlowOption {
	return func(ic *flowconfig) {
		ic.stringSource = func(errors chan<- error) <-chan string {
			output := make(chan string, 1)
			output <- input
			close(output)
			return output
		}
	}
}

// CSVFirstFieldOnlySource sets a CSV reader as source and emit only the
// first field/column (or the entire line if the input isn't a proper CSV)
func CSVFirstFieldOnlySource(input io.Reader) FlowOption {
	return func(ic *flowconfig) {
		ic.stringSource = func(errors chan<- error) <-chan string {
			output := make(chan string)
			go func() {
				reader := csv.NewReader(input)
				for {
					record, err := reader.Read()
					if err != nil {
						if err == io.EOF {
							close(output)
							return
						}
						errors <- fmt.Errorf("csv read error: %w", err)
					} else {
						output <- record[0]
					}
				}
			}()
			return output
		}
	}
}

// CSVSink sets a CSV writer as sink. The columns param selects the fields of Item
// to be written in the CSV output.
func CSVSink(output io.Writer, columns *[]string) FlowOption {
	return func(ic *flowconfig) {
		ic.itemSink = func(errors chan<- error, input <-chan Item) {
			w := csv.NewWriter(output)
			for item := range input {
				err := w.Write(item.AsRecord(columns))
				if err != nil {
					// FIXME fatal?
					errors <- fmt.Errorf("csv write error: %w", err)
				}
			}
			w.Flush()
			if w.Error() != nil {
				errors <- fmt.Errorf("csv write error: %w", w.Error())
			}
		}
	}
}

// TableSink sets a text/tabwriter as sink for a human readable output.
// The columns param selects the fields of Item to be written in the output.
func TableSink(output io.Writer, columns *[]string) FlowOption {
	return func(ic *flowconfig) {
		ic.itemSink = func(errors chan<- error, input <-chan Item) {
			w := tabwriter.NewWriter(output, 0, 8, 2, ' ', 0)
			for item := range input {
				_, err := fmt.Fprintln(w, strings.Join(item.AsRecord(columns), "\t"))
				if err != nil {
					// FIXME fatal?
					errors <- fmt.Errorf("table write error: %w", err)
				}
			}
			err := w.Flush()
			if err != nil {
				errors <- fmt.Errorf("table write error: %w", err)
			}
		}
	}
}

// NullSink sets a discard sink.
func NullSink() FlowOption {
	return func(ic *flowconfig) {
		ic.itemSink = func(errors chan<- error, input <-chan Item) {
			for range input {
				// noop
			}
		}
	}
}

// JSONLinesSink sets a JSON Lines as sink.
func JSONLinesSink(output io.Writer) FlowOption {
	return func(ic *flowconfig) {
		ic.itemSink = func(errors chan<- error, input <-chan Item) {
			enc := json.NewEncoder(output)
			for i := range input {
				if err := enc.Encode(i); err != nil {
					errors <- fmt.Errorf("jsonl write error: %w", err)
				}
			}
		}
	}
}

func options2flowconfig(cfgs ...FlowOption) flowconfig {
	var cfg flowconfig
	for _, c := range cfgs {
		c(&cfg)
	}
	return cfg
}
