package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Output struct {
	JSON bool
	Out  io.Writer
	Err  io.Writer
	In   io.Reader
}

func NewOutput(jsonOutput bool) Output {
	return Output{JSON: jsonOutput, Out: os.Stdout, Err: os.Stderr, In: os.Stdin}
}

func (o Output) PrintJSON(payload any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return err
	}
	_, err := fmt.Fprint(o.Out, buf.String())
	return err
}
