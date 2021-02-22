package main

import (
	"github.com/kuberlogic/operator/cmd"
	"os"
)

func main() {
	cmd.Main(os.Args[1:])
}
