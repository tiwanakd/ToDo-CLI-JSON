package main

import (
	"fmt"
	"os"

	"github.com/tiwanakd/ToDo-CLI-JSON/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

}
