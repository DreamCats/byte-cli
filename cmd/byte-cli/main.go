package main

import (
	"os"

	"github.com/DreamCats/byte-cli/internal/cli"
)

func main() {
	cli.Main(os.Args[1:])
}
