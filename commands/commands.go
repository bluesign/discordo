package commands

import (
	"errors"
	"sort"
	"strings"
	"unicode"

	"github.com/bluesign/discordo/widgets"
	"github.com/google/shlex"
)

type Command interface {
	Aliases() []string
	Execute(*widgets.Aerc, []string) error
	Complete(*widgets.Aerc, []string) []string
}

type Commands map[string]Command

func NewCommands() *Commands {
	cmds := Commands(make(map[string]Command))
	return &cmds
}

func (cmds *Commands) dict() map[string]Command {
	return map[string]Command(*cmds)
}

func (cmds *Commands) Names() []string {
	names := make([]string, 0)

	for k := range cmds.dict() {
		names = append(names, k)
	}
	return names
}

func (cmds *Commands) Register(cmd Command) {
	// TODO enforce unique aliases, until then, duplicate each
	if len(cmd.Aliases()) < 1 {
		return
	}
	for _, alias := range cmd.Aliases() {
		cmds.dict()[alias] = cmd
	}
}

type NoSuchCommand string

func (err NoSuchCommand) Error() string {
	return "Unknown command " + string(err)
}

type CommandSource interface {
	Commands() *Commands
}

func (cmds *Commands) ExecuteCommand(aerc *widgets.Aerc, args []string) error {
	if len(args) == 0 {
		return errors.New("Expected a command.")
	}
	if cmd, ok := cmds.dict()[args[0]]; ok {
		return cmd.Execute(aerc, args)
	}
	return NoSuchCommand(args[0])
}

func (cmds *Commands) GetCompletions(aerc *widgets.Aerc, cmd string) []string {
	args, err := shlex.Split(cmd)
	if err != nil {
		return nil
	}

	if len(args) == 0 {
		names := cmds.Names()
		sort.Strings(names)
		return names
	}

	if len(args) > 1 || cmd[len(cmd)-1] == ' ' {
		if cmd, ok := cmds.dict()[args[0]]; ok {
			var completions []string
			if len(args) > 1 {
				completions = cmd.Complete(aerc, args[1:])
			} else {
				completions = cmd.Complete(aerc, []string{})
			}
			if completions != nil && len(completions) == 0 {
				return nil
			}

			options := make([]string, 0)
			for _, option := range completions {
				options = append(options, args[0]+" "+option)
			}
			return options
		}
		return nil
	}

	names := cmds.Names()
	options := make([]string, 0)
	for _, name := range names {
		if strings.HasPrefix(name, args[0]) {
			options = append(options, name)
		}
	}

	if len(options) > 0 {
		return options
	}
	return nil
}

// CompletionFromList provides a convenience wrapper for commands to use in the
// Complete function. It simply matches the items provided in valid
func CompletionFromList(valid []string, args []string) []string {
	out := make([]string, 0)
	if len(args) == 0 {
		return valid
	}
	for _, v := range valid {
		if hasCaseSmartPrefix(v, args[0]) {
			out = append(out, v)
		}
	}
	return out
}

func GetLabels(aerc *widgets.Aerc, args []string) []string {

	out := make([]string, 0)

	return out
}

// hasCaseSmartPrefix checks whether s starts with prefix, using a case
// sensitive match if and only if prefix contains upper case letters.
func hasCaseSmartPrefix(s, prefix string) bool {
	if hasUpper(prefix) {
		return strings.HasPrefix(s, prefix)
	}
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}

func hasUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}
