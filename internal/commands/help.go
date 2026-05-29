package commands

import "fmt"

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			return true
		}
	}
	return false
}

func printHelp(out Output, text string) {
	fmt.Fprintln(out.Out, text)
}
