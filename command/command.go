package command

import (
	log "github.com/sirupsen/logrus"
	"flag"
	"fmt"
)

const COMMAND_RUN  = "run"
const COMMAND_INIT = "init"

type Command struct {
	Usage string
	FlagSet *flag.FlagSet
	Flags map[string]interface{}
	Action func(map[string]interface{}, []string) error
}

var runFlagSet = flag.NewFlagSet(COMMAND_RUN, flag.ExitOnError)
var runCmd = Command{
	Usage: "Create a container with namespace and cgroups limit",
	FlagSet: runFlagSet,
	Flags: map[string]interface{}{
		"tty": runFlagSet.Bool("tty", false, "enable tty"),
	},
	Action: func(argKV map[string]interface{}, tail []string) error {
		if len(tail) < 1 {
			return fmt.Errorf("Missing container command")
		}
		cmd := tail[0]
		log.Info("command %s", cmd)
		// tty := argKV["tty"].(bool)
		// Run(tty, cmd)
		return nil
	},
}

var initFlagSet = flag.NewFlagSet(COMMAND_INIT, flag.ExitOnError)
var initCmd = Command{
	Usage: "Init container process run user's process in container. Do not call it outside",
	FlagSet: initFlagSet,
	Flags: map[string]interface{}{},
	Action: func(argKV map[string]interface{}, tail []string) error {
		if len(tail) < 1 {
			return fmt.Errorf("Missing container init process command")
		}
		cmd := tail[0]
		log.Info("command %s", cmd)
		// err := container.RunContainerInitProcess(cmd)
		// return err
		return nil
	},
}
