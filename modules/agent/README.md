## Overview
Kuberlogic Agent is a library that allows scheduling of local command execution for Kubernetes pods using sidecar container pattern.

## How it works
Kuberlogic Agent is a combination of two logical entities - `AgentController` and `Agent`. `AgentController` is responsible for commands scheduling and command failures processing, while `Agent` only executes commands and notifies `AgentController` about their statuses.

Internal communication is done using `grpc` protocol, and it is initiated by an `Agent`.

## Usage
### Agent

Running agent is this easy:
```
import (
	"github.com/kuberlogic/operator/modules/agent/client"
	"os"
	"log"
)

func main() {
	c, err := client.NewClient(
		os.Getenv("NAME"),
		os.Getenv("CONTROLLER_ADDR"),
		os.Getenv("INIT_ONLY") != "")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(c.Run())
}
```

Few more notes:
* `Agent` can’t be run without `AgentController`!
* `Agent` can be run in `init` mode, where it only executes init commands and then exits.
* Make sure that `Agent` name is unique for `AgentController`!

## AgentController
`AgentController` is designed to be run as a part of a bigger application:
```
import (
	"github.com/kuberlogic/operator/modules/agent/controller"
)

func main() {
	controller.Init(18888) // controller port; After this call controller will be started and run in another goroutine
}
```

## Scheduling commands
```
import (
	"github.com/kuberlogic/operator/modules/agent/controller"
	"github.com/kuberlogic/operator/modules/agent/command"
)

// some code ommited. we assume controller is already 
	testCommand, err := command.NewCommand("echo", "echo", 	"forever", "whatever")
	if err != nil {
		log.Fatal(err)
	}
	controller.Controller.Commands.Add("test", testCommand)
}
```

## Build / Release
```
make br VERSION=<semver>
```