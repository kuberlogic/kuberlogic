package client

import (
	"context"
	agentgrpc "github.com/kuberlogic/operator/modules/agent/agent-grpc"
	"github.com/kuberlogic/operator/modules/agent/command"
	"google.golang.org/grpc"
	"log"
	"time"
)

type AgentClient struct {
	name             string
	controllerClient agentgrpc.CommandAPIClient

	initOnly bool
}

func (a *AgentClient) GetCommand(ctx context.Context, commandType int32) (*command.Command, error) {
	req := &agentgrpc.CommandRequest{
		AgentName:   a.name,
		CommandType: commandType,
	}

	log.Printf("sending command request: %v\n", req)
	c, err := a.controllerClient.GetCommand(ctx, req)
	if err != nil {
		return nil, err
	}
	log.Printf("found command: %v\n", c)
	return &command.Command{
		Name:    c.CommandName,
		Command: c.Command,
		Args:    c.Args,
	}, nil
}

func (a *AgentClient) ExecuteControllerCommand(ctx context.Context, c *command.Command, cType int32) error {
	res := &agentgrpc.CommandResult{
		AgentName:     a.name,
		CommandName:   c.Name,
		CommandType:   cType,
		CommandResult: command.SuccessCode,
	}

	if errCmd := c.Execute(); errCmd != nil {
		res.CommandResult = command.FailureCode
	}
	log.Printf("command execution result: %v\n", res)
	_, err := a.controllerClient.CommandExecutionResult(ctx, res)
	return err
}

func (a *AgentClient) ExecInitCommand() error {
	ctx := context.TODO()
	initCom, err := a.GetCommand(ctx, command.InitCommandType)
	if err != nil {
		return err
	}
	if initCom.Empty() {
		return nil
	}

	return a.ExecuteControllerCommand(ctx, initCom, command.InitCommandType)

}

func (a *AgentClient) ExecRuntimeCommand() error {
	ctx := context.TODO()
	runtimeCom, err := a.GetCommand(ctx, command.RuntimeCommandType)
	if err != nil {
		return err
	}
	if runtimeCom.Empty() {
		return nil
	}
	return a.ExecuteControllerCommand(ctx, runtimeCom, command.RuntimeCommandType)
}

func (a *AgentClient) Run() error {
	log.Println("asking for init command")
	if err := a.ExecInitCommand(); err != nil {
		return err
	}
	if a.initOnly {
		return nil
	}

	for {
		log.Println("asking for runtime command")
		if err := a.ExecRuntimeCommand(); err != nil {
			return err
		}
		time.Sleep(time.Second * 60)
	}
}

func NewClient(name, controllerAddr string, initOnly bool) (*AgentClient, error) {
	conn, err := grpc.Dial(controllerAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := agentgrpc.NewCommandAPIClient(conn)

	return &AgentClient{
		name:             name,
		controllerClient: c,
		initOnly:         initOnly,
	}, nil
}
