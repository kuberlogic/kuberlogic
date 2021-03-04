package main

import (
	"github.com/kuberlogic/operator/modules/apiserver/cmd"
	"os"
)

func main() {
	cmd.Main(os.Args[1:])
}
