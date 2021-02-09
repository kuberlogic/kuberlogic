package main

import (
	"gitlab.com/cloudmanaged/operator/cmd"
	"os"
)

func main() {
	cmd.Main(os.Args[1:])
}
