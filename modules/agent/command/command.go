/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	Name    string
	Command string
	Args    []string
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
