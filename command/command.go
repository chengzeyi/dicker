package command

import (
	"flag"
	"fmt"

	"github.com/chengzeyi/dicker/container"
	log "github.com/sirupsen/logrus"
)

const COMMAND_RUN  = "run"
const COMMAND_INIT = "init"

type Command struct {
	Usage string
	FlagSet *flag.FlagSet
	Flags map[string]interface{}
	Action func(map[string]interface{}, []string) error
}

// Parse a subcommand's arguments, excluding the name of the command.
func (c *Command) Execute(args []string) error {
	if err := c.FlagSet.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprintf(c.FlagSet.Output(), "%s: %s\n", c.Name(), c.Usage)
			c.FlagSet.PrintDefaults()
		} else {
			return fmt.Errorf("Parse command line flags error %v", err)
		}
	}

	// This could be redundant.
	if !c.FlagSet.Parsed() {
		return fmt.Errorf("Commnd %s has not been parsed", c.Name())
	}

	tail := c.FlagSet.Args()
	if err := c.Action(c.Flags, tail); err != nil {
		return fmt.Errorf("Do Action of command %s error %v", c.Name(), err)
	}

	return nil
}

func (c *Command) Name() string {
	return c.FlagSet.Name()
}

var runFlagSet = flag.NewFlagSet(COMMAND_RUN, flag.ContinueOnError)
var runCmd = Command{
	Usage: "Create a container with namespace and cgroups limit, [OPTION]... <IMAGE> <COMMAND> [ARG]...",
	FlagSet: runFlagSet,
	Flags: map[string]interface{}{
		"tty": runFlagSet.Bool("tty", false, "enable tty"),
	},
	Action: func(argKV map[string]interface{}, tail []string) error {
		if len(tail) == 0 {
			return fmt.Errorf("Missing container image")
		}
		if len(tail) == 1 {
			return fmt.Errorf("Missing container command")
		}
		imageName := tail[0]
		cmdArr := tail[1:]
		log.Info("image name %s, command array %v", imageName, cmdArr)
		runOption := &RunOption{
			Tty: argKV["tty"].(bool),
		}
		if err := Run(runOption, imageName, cmdArr); err != nil {
			return fmt.Errorf("Run image %s and command array %v error %v", imageName, cmdArr, err)
		}

		return nil
	},
}

var initFlagSet = flag.NewFlagSet(COMMAND_INIT, flag.ContinueOnError)
var initCmd = Command{
	Usage: "Init container process and run user's process in container. Do not call it outside",
	FlagSet: initFlagSet,
	Flags: map[string]interface{}{},
	Action: func(argKV map[string]interface{}, tail []string) error {
		if len(tail) > 0 {
			log.Warnf("Init process only reads command from pipe with parent process.")
		}

		if err := container.RunContainerInitProcess(); err != nil {
			return fmt.Errorf("RunContainerInitProcess error %v", err)
		}

		return nil
	},
}
