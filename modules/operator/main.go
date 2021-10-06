package main

import (
	"github.com/kuberlogic/kuberlogic/modules/operator/cmd"
	"os"
)

func main() {
	cmd.Main(os.Args[1:])
}
