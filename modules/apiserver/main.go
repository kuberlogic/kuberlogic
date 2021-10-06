package main

import (
	"github.com/kuberlogic/kuberlogic/modules/apiserver/cmd"
	"os"
)

func main() {
	cmd.Main(os.Args[1:])
}
