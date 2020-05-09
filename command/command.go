package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chengzeyi/dicker/container"
	log "github.com/sirupsen/logrus"
)

const COMMAND_HELP = "help"
const COMMAND_RUN  = "run"
const COMMAND_INIT = "init"

type ICommand interface {
	Execute(args []string) error
	Name() string
	Usage() string
	Help()
}

var commandMap = map[string] ICommand{}

func init() {
	commandMap[COMMAND_HELP] = &helpCmd
	commandMap[COMMAND_RUN] = &runCmd
	commandMap[COMMAND_INIT] = &initCmd
}

func GetCommand(cmdName string) ICommand {
	return commandMap[cmdName]
}

type Command struct {
	usage string
	flagSet *flag.FlagSet
	flags map[string]interface{}
	action func(map[string]interface{}, []string) error
}

// Print help information as name: usage\n [flag]...
func (c *Command) Help() {
	fmt.Fprintf(os.Stderr, "%s: %s\n", c.Name(), c.Usage())
	if c.flagSet != nil {
		c.flagSet.PrintDefaults()
	}
	fmt.Fprintln(os.Stderr)
}

func (c *Command) Usage() string {
	return c.usage
}

// Parse a subcommand's arguments, excluding the name of the command.
func (c *Command) Execute(args []string) error {
	if err := c.flagSet.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprintf(c.flagSet.Output(), "%s: %s\n", c.Name(), c.Usage())
			c.flagSet.PrintDefaults()
		} else {
			return fmt.Errorf("Parse command line flags error %v", err)
		}
	}

	// This could be redundant.
	if !c.flagSet.Parsed() {
		return fmt.Errorf("Commnd %s has not been parsed", c.Name())
	}

	tail := c.flagSet.Args()
	if err := c.action(c.flags, tail); err != nil {
		return fmt.Errorf("Do action of command %s error %v", c.Name(), err)
	}

	return nil
}

func (c *Command) Name() string {
	return c.flagSet.Name()
}

var helpCmd = Command{
	usage: "Look up help for commands, [COMMAND]...",
	flagSet: &flag.FlagSet{},
	flags: map[string]interface{}{},
	action: func(argKV map[string]interface{}, tail []string) error {
		if len(tail) == 0 {
			for k, v := range commandMap {
				fmt.Fprintf(os.Stderr, "%s\t%s\n", k, v.Usage())
			}
		} else {
			for _, v := range tail {
				if GetCommand(v) == nil {
					return fmt.Errorf("Unknown Dicker command %s", v)
				}
			}
			for _, v := range tail {
				GetCommand(v).Help()
			}
		}

		return nil
	},
}

var runFlagSet = flag.NewFlagSet(COMMAND_RUN, flag.ContinueOnError)
var runCmd = Command{
	usage: "Create a container with namespace and cgroups limit, [OPTION]... <IMAGE> <COMMAND> [ARG]...",
	flagSet: runFlagSet,
	flags: map[string]interface{}{
		"tty": runFlagSet.Bool("tty", false, "enable tty"),
		"container-name": runFlagSet.String("container-name", "", "set container name"),
		"volume-mapping": runFlagSet.String("volume-mapping", "", "':' delimited mapping to mount a host volume to a container volume"),
		"envs": runFlagSet.String("environments", "", "':' delimited environment variables"),
	},
	action: func(argKV map[string]interface{}, tail []string) error {
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
			ContainerName: argKV["container-name"].(string),
			VolumeMapping: argKV["volume-mapping"].(string),
			Envs: strings.Split(argKV["envs"].(string), ":"),
		}
		if err := Run(runOption, imageName, cmdArr); err != nil {
			return fmt.Errorf("Run image %s and command array %v error %v", imageName, cmdArr, err)
		}

		return nil
	},
}

var initFlagSet = flag.NewFlagSet(COMMAND_INIT, flag.ContinueOnError)
var initCmd = Command{
	usage: "Init container process and run user's process in container. Do not call it outside",
	flagSet: initFlagSet,
	flags: map[string]interface{}{},
	action: func(argKV map[string]interface{}, tail []string) error {
		if len(tail) > 0 {
			log.Warnf("Init process only reads command from pipe with parent process.")
		}

		if err := container.RunContainerInitProcess(); err != nil {
			return fmt.Errorf("RunContainerInitProcess error %v", err)
		}

		return nil
	},
}
