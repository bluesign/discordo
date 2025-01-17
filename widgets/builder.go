package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/ayntgl/astatine"
	"github.com/bluesign/discordo/discord"
)

func BuildMessage(Session *astatine.Session, m *astatine.Message, old *astatine.Message) []byte {
	return buildMessage(Session, m, old)
}

func buildMessage(Session *astatine.Session, m *astatine.Message, old *astatine.Message) []byte {
	var b strings.Builder

	switch m.Type {
	case astatine.MessageTypeDefault, astatine.MessageTypeReply:
		// Define a new region and assign message ID as the region ID.
		// Learn more:
		// https://pkg.go.dev/github.com/rivo/tview#hdr-Regions_and_Highlights
		b.WriteString("[\"")
		b.WriteString(m.ID)
		b.WriteString("\"]")

		if old == nil || old.Author.Username != m.Author.Username {
			b.WriteString("\n")
		}

		// Build the author of this message.
		if old == nil || (m.ReferencedMessage != nil) {
			buildReferencedMessage(&b, m.ReferencedMessage, Session.State.User.ID)
		}

		if old == nil || old.Author.Username != m.Author.Username {
			pre := ""
			if m.Thread != nil {
				pre = fmt.Sprintf(" (thread %d messages) ", m.Thread.MessageCount)
			}
			reactions := ""
			for _, react := range m.Reactions {
				reactions = fmt.Sprintf("%s %d %s ", reactions, react.Count, react.Emoji.Name)
			}

			b.WriteString(buildAuthor(m.Author, Session.State.User.ID, fmt.Sprintf("%s %s %s", pre, m.Timestamp.Format(time.Stamp), reactions)))
			b.WriteString("\n")
		}

		/*b.WrteString("[::d]")
		b.WriteString(m.Timestamp.Format(time.Stamp))
		b.WriteString("[::-]")
		b.WriteByte('\n')*/

		// Build the message associated with crosspost, channel follow add, pin, or a reply.

		// Build the contents of the message.

		buildContent(&b, m, Session.State.User.ID)

		if m.EditedTimestamp != nil {
			b.WriteString(" [::d](edited)[::-]")
		}

		// Build the embeds associated with the message.
		buildEmbeds(&b, m.Embeds)

		// Build the message attachments (attached files to the message).
		buildAttachments(&b, m.Attachments)

		// Tags with no region ID ([""]) do not start new regions. They can
		// therefore be used to mark the end of a region.
		b.WriteString("[\"\"]")

		b.WriteString("\n")

	case astatine.MessageTypeGuildMemberJoin:
		b.WriteString("[#5865F2]")
		b.WriteString(m.Author.Username)
		b.WriteString("[-] joined the server.")

		b.WriteByte('\n')
	case astatine.MessageTypeCall:
		b.WriteString("[#5865F2]")
		b.WriteString(m.Author.Username)
		b.WriteString("[-] started a call.")

		b.WriteByte('\n')
	case astatine.MessageTypeChannelPinnedMessage:
		b.WriteString("[#5865F2]")
		b.WriteString(m.Author.Username)
		b.WriteString("[-] pinned a message.")

		b.WriteByte('\n')
	}

	if str := b.String(); str != "" {
		b := make([]byte, len(str)+1)
		copy(b, str)

		return b
	}

	return nil
}

func buildReferencedMessage(b *strings.Builder, rm *astatine.Message, clientID string) {
	if rm != nil {
		b.WriteString("[::d] > ")
		b.WriteString(buildAuthor(rm.Author, clientID, ""))
		b.WriteString("[::d] ")

		if rm.Content != "" {
			rm.Content = buildMentions(rm.Content, rm.Mentions, clientID)
			str := discord.ParseMarkdown(rm.Content)
			if len(str) > 700 {
				str = str[:700]
			}
			b.WriteString(str)
		}

		b.WriteString("[::-]")
		b.WriteString("[::-]")
		b.WriteByte('\n')
	}
}

func buildContent(b *strings.Builder, m *astatine.Message, clientID string) {

	if m.Content != "" {
		m.Content = buildMentions(m.Content, m.Mentions, clientID)

		parsed := discord.ParseMarkdown(m.Content)

		b.WriteString(parsed)

		//w.Write([]byte(img.Render()))

	}

}

func buildEmbeds(b *strings.Builder, es []*astatine.MessageEmbed) {
	for _, e := range es {
		if e.Type != astatine.EmbedTypeRich {
			continue
		}

		var (
			embedBuilder strings.Builder
			hasHeading   bool
		)
		prefix := fmt.Sprintf("[#%06X]▐[-] ", e.Color)

		b.WriteByte('\n')
		embedBuilder.WriteString(prefix)

		if e.Author != nil {
			hasHeading = true
			embedBuilder.WriteString("[::u]")
			embedBuilder.WriteString(e.Author.Name)
			embedBuilder.WriteString("[::-]")
		}

		if e.Title != "" {
			if hasHeading {
				embedBuilder.WriteByte('\n')
				embedBuilder.WriteByte('\n')
			}

			embedBuilder.WriteString("[::b]")
			embedBuilder.WriteString(e.Title)
			embedBuilder.WriteString("[::-]")
		}

		if e.Description != "" {
			if hasHeading {
				embedBuilder.WriteByte('\n')
				embedBuilder.WriteByte('\n')
			}

			embedBuilder.WriteString(discord.ParseMarkdown(e.Description))
		}

		if len(e.Fields) != 0 {
			if hasHeading || e.Description != "" {
				embedBuilder.WriteByte('\n')
				embedBuilder.WriteByte('\n')
			}

			for i, ef := range e.Fields {
				embedBuilder.WriteString("[::b]")
				embedBuilder.WriteString(ef.Name)
				embedBuilder.WriteString("[::-]")
				embedBuilder.WriteByte('\n')
				embedBuilder.WriteString(discord.ParseMarkdown(ef.Value))

				if i != len(e.Fields)-1 {
					embedBuilder.WriteString("\n\n")
				}
			}
		}

		if e.Footer != nil {
			if hasHeading {
				embedBuilder.WriteString("\n\n")
			}

			embedBuilder.WriteString(e.Footer.Text)
		}

		b.WriteString(strings.ReplaceAll(embedBuilder.String(), "\n", "\n"+prefix))
	}
}

func buildAttachments(b *strings.Builder, as []*astatine.MessageAttachment) {
	for _, a := range as {

		b.WriteByte('\n')
		b.WriteByte('[')
		b.WriteString(a.Filename)
		b.WriteString("]: ")
		b.WriteString(a.URL)
		/*
			if strings.HasSuffix(a.Filename, ".png") {

				var img *ansimage.ANSImage
				img, _ = ansimage.NewScaledFromURL(a.URL, 60, 60, color.Black, ansimage.ScaleModeFit, ansimage.NoDithering)
				w := tview.ANSIWriter(b)
				w.Write([]byte(img.Render()))
			} else {
				b.WriteByte('\n')
				b.WriteByte('[')
				b.WriteString(a.Filename)
				b.WriteString("]: ")
				b.WriteString(a.URL)

			}*/

	}
}

func buildMentions(content string, mentions []*astatine.User, clientID string) string {
	for _, mUser := range mentions {
		var color string
		if mUser.ID == clientID {
			color = "[:#5865F2]"
		} else {
			color = "[#EB459E]"
		}

		content = strings.NewReplacer(
			// <@USER_ID>
			"<@"+mUser.ID+">",
			color+"@"+mUser.Username+"[-:-]",
			// <@!USER_ID>
			"<@!"+mUser.ID+">",
			color+"@"+mUser.Username+"[-:-]",
		).Replace(content)
	}

	return content
}

func buildAuthor(u *astatine.User, clientID string, timeStamp string) string {
	color := "[#ED4245]"
	if u.ID == clientID {
		color = "[#57F287]"
	}

	msg := fmt.Sprintf("%s%s[-][::d] %s[::-]", color, u.Username, timeStamp)

	return msg

}
