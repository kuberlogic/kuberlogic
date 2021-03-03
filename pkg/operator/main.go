package main

import (
	"github.com/kuberlogic/operator/pkg/operator/cmd"
	"os"
)

func main() {
	cmd.Main(os.Args[1:])
}
