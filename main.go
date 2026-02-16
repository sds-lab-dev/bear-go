package main

import (
	"fmt"
	"os"

	"github.com/sds-lab-dev/bear-go/app"
)

func main() {
	if err := app.Run(os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
