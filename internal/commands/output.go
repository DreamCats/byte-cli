package commands

import (
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
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(o.Out, string(data))
	return err
}
