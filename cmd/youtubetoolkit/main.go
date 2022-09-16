package main

import (
	"fmt"

	"github.com/raffaelecassia/youtubetoolkit/cmd/youtubetoolkit/commands"
)

func main() {
	err := commands.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
