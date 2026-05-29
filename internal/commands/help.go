package commands

import (
	"fmt"
	"strings"
)

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

func normalizeFlags(args []string, valueFlags, boolFlags map[string]bool) []string {
	flags := []string{}
	positionals := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i:]...)
			break
		}
		name := flagName(arg)
		if name == "" {
			positionals = append(positionals, arg)
			continue
		}
		if boolFlags[name] {
			flags = append(flags, arg)
			continue
		}
		if valueFlags[name] {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positionals = append(positionals, arg)
	}
	return append(flags, positionals...)
}

func flagName(arg string) string {
	if !strings.HasPrefix(arg, "-") || arg == "-" {
		return ""
	}
	name := strings.TrimLeft(arg, "-")
	if idx := strings.IndexByte(name, '='); idx >= 0 {
		name = name[:idx]
	}
	return name
}

func stringSet(values ...string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}
