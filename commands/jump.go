package commands

import (
	"errors"
	"regexp"
	"strings"

	"github.com/ayntgl/astatine"
	"github.com/bluesign/discordo/config"
	"github.com/bluesign/discordo/lib/ui"
	"github.com/bluesign/discordo/widgets"
)

var ()

type ChangeDirectory struct{}

func init() {
	register(ChangeDirectory{})
}

func (ChangeDirectory) Aliases() []string {
	return []string{"jump"}
}

func (ChangeDirectory) Complete(aerc *widgets.Aerc, args []string) []string {
	input := strings.Join(args, " ")

	var channels []string
	for _, guild := range aerc.Session.State.Guilds {
		for _, channel := range guild.Channels {

			perm, _ := aerc.Session.State.UserChannelPermissions(aerc.Session.State.User.ID, channel.ID)

			if (channel.IsThread()) || (perm&astatine.PermissionViewChannel > 0 && strings.Contains(config.Channels(), channel.ID+"\n")) {
				if strings.Contains(channel.Name, input) {
					re, _ := regexp.Compile(`[^a-z|^A-Z|^0-9|^_|^-]`)
					cleanName := re.ReplaceAllString(channel.Name, "")
					channels = append(channels, cleanName)
				}
			}

		}
	}

	return channels
}

func SelectNodeByLabel(label string, root *ui.TreeNode, tree *ui.TreeView) bool {
	for _, child := range root.GetChildren() {

		if child.GetText() == label {
			tree.SetCurrentNode(child)
			return false
		}
		if !SelectNodeByLabel(label, child, tree) {
			return false
		}
	}

	return true
}

func (ChangeDirectory) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) < 1 {
		return errors.New("Usage: jump [ channel | DM ]")
	}

	SelectNodeByLabel(strings.Join(args[1:], " "), aerc.AccountView.ChannelTree.GetRoot(), aerc.AccountView.ChannelTree)
	aerc.AccountView.Invalidate()
	return nil
}
