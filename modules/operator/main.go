package main

import (
	"github.com/kuberlogic/operator/modules/operator/cmd"
	"os"
)

func main() {
	cmd.Main(os.Args[1:])
}
