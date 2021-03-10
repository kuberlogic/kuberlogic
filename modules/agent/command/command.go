package command

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const (
	InitCommandType = iota + 1
	RuntimeCommandType
)
const (
	SuccessCode = iota + 1
	FailureCode
)

type Command struct {
	Name string
	Command string
	Args []string
}

type Queue struct {
	Q map[string]*Command
}

func (c *Command) Execute() error {
	log.Printf("executing command: %v\n", c)
	com := exec.Command(c.Command, c.Args...)
	com.Stdout, com.Stderr = os.Stdout, os.Stderr

	return com.Run()
}

func (c *Command) Empty() bool {
	return c == nil || c.Name == ""
}

func NewCommand(name, command string, args ...string) (*Command, error) {
	if name == "" || command == "" {
		return nil, fmt.Errorf("name or command can't be empty")
	}
	return &Command{
		Name:    name,
		Command: command,
		Args:    args,
	}, nil
}

func (q *Queue) Get(agent string) *Command {
	c, _ := q.Q[agent]
	return c
}

func (q *Queue) Add(agent string, c *Command) {
	q.Q[agent] = c
}

func (q *Queue) Del(agent string) {
	delete(q.Q, agent)
}

func NewCommandQ() *Queue {
	return &Queue{
		Q: make(map[string]*Command, 0),
	}
}