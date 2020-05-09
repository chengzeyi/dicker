package main

import (
	"os"
	"github.com/chengzeyi/dicker/command"

	log "github.com/sirupsen/logrus"
)

const usage = "Dicker is a simple container runtime implementation. Use it just for fun."

func main() {
	args := os.Args
	if len(args) <= 1 {
		log.Errorf("Missing Dicker command, use help command to see the full command list and usage")
		os.Exit(1)
	}

	cmdName := os.Args[1]
	cmd := command.GetCommand(cmdName)
	if cmd == nil {
		log.Errorf("Unknown Dicker command %s", cmdName)
		os.Exit(1)
	}

	if err := cmd.Execute(args[2:]); err != nil {
		log.Errorf("Execute Dicker command %s error %v", cmdName, err)
		os.Exit(1)
	}
}
