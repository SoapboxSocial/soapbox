package main

import (
	"os"

	"github.com/soapboxsocial/soapbox/cmd/accounts/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
