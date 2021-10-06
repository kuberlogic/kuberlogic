package controller

import (
	"context"
	"fmt"
	agentgrpc "github.com/kuberlogic/kuberlogic/modules/agent/agent-grpc"
	"github.com/kuberlogic/kuberlogic/modules/agent/command"
	"google.golang.org/grpc"
	"log"
	"net"
)

type AgentController struct {
	port         int
	InitCommands *command.Queue
	Commands     *command.Queue
}

func (a *AgentController) AddRuntimeCommand(agent string, com string, args ...string) error {
	c, err := command.NewCommand(agent+com, com, args...)
	if err != nil {
		return err
	}
	a.Commands.Add(agent, c)
	return nil
}

// processes command request
// gets command from queue
func (a *AgentController) GetCommand(ctx context.Context, request *agentgrpc.CommandRequest) (*agentgrpc.Command, error) {
	log.Printf("got command request: %v\n", request)
	c := &command.Command{}

	if request.AgentName == "" {
		return nil, fmt.Errorf("agent name can't be empty")
	}

	switch request.CommandType {
	case command.InitCommandType:
		c = a.InitCommands.Get(request.AgentName)
	case command.RuntimeCommandType:
		c = a.Commands.Get(request.AgentName)
	default:
		return nil, fmt.Errorf("unknown command type %d", request.CommandType)
	}

	if c.Empty() {
		return &agentgrpc.Command{}, nil
	}

	return &agentgrpc.Command{
		CommandName: c.Name,
		Command:     c.Command,
		Args:        c.Args,
	}, nil
}

// processes command execution result
func (a *AgentController) CommandExecutionResult(ctx context.Context, result *agentgrpc.CommandResult) (*agentgrpc.CommandResultAck, error) {
	log.Printf("processing command result: %v\n", result)
	if result.AgentName == "" || result.CommandName == "" {
		return nil, fmt.Errorf("agent or command name can't be empty")
	}
	ret := &agentgrpc.CommandResultAck{
		CommandName: result.CommandName,
		Accepted:    true,
	}

	if result.CommandResult != command.SuccessCode {
		ret.Retry = true
		return ret, nil
	}

	switch result.CommandType {
	case command.InitCommandType:
		a.InitCommands.Del(result.AgentName)
	case command.RuntimeCommandType:
		a.Commands.Del(result.AgentName)
	default:
		ret.Accepted = false
	}
	return ret, nil
}

func (a *AgentController) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	agentgrpc.RegisterCommandAPIServer(s, a)

	log.Printf("starting grpc server: %v\n", s)
	log.Fatal(s.Serve(lis))
}

func NewController(port int) *AgentController {
	a := &AgentController{
		port:         port,
		InitCommands: command.NewCommandQ(),
		Commands:     command.NewCommandQ(),
	}
	return a
}
