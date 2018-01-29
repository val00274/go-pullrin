package pullrin

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/nlopes/slack"
	"golang.org/x/net/context"
)

type NotificationChannel struct {
	client  *slack.Client
	channel string
}

func OpenNotificationChannel(token, channel string) *NotificationChannel {
	ch := &NotificationChannel{}
	ch.client = slack.New(token)
	ch.channel = channel
	return ch
}

func (c *NotificationChannel) Post(user, icon, message string, attachments []slack.Attachment) {
	c.client.PostMessage(c.channel, message, slack.PostMessageParameters{Username: user, IconURL: icon, Attachments: attachments})
}

type NotificationItem struct {
	repo      *Repository
	pr        *github.PullRequest
	reactions map[string]string
	reviewers []string
}

func NewNotificationItem(repo *Repository,
	pr *github.PullRequest,
	reactions map[string]string,
	reviewers []string) *NotificationItem {
	return &NotificationItem{repo, pr, reactions, reviewers}
}

func (m *NotificationItem) IsAuthor(name string) bool {
	return name == *m.pr.User.Login
}

func (m *NotificationItem) IconName(name string) string {
	if m.IsAuthor(name) {
		return "god"
	}

	switch m.reactions[name] {
	case "+1":
		return "+1"
	case "-1":
		return "-1"
	case "laugh":
		return "smile"
	case "confused":
		return "confused"
	case "heart":
		return "heart"
	case "hooray":
		return "tada"
	default:
		return "space"
	}
}

func (m *NotificationItem) IsComplete() bool {
	return !strings.Contains(m.ReactionTable(), ":space:")
}

func (m *NotificationItem) Color() (color string) {
	if m.IsComplete() {
		color = "good"
	} else {
		color = "warning"
	}
	return
}

func (m *NotificationItem) Title() (title string) {
	ctx := context.Background()

	var balloonMark string
	if n := m.repo.FetchCommentCount(ctx, m.pr); n > 0 {
		balloonMark = fmt.Sprintf(" :speech_balloon: %d", n)
	} else {
		balloonMark = ""
	}

	title = fmt.Sprintf("#%d %s%s", *m.pr.Number, *m.pr.Title, balloonMark)
	return
}

func (m *NotificationItem) ReactionTable() string {
	var buf bytes.Buffer
	for _, name := range m.reviewers {
		if m.IsAuthor(name) {
			buf.WriteString(fmt.Sprintf("*%s* :%s: ", name, m.IconName(name)))
		} else {
			buf.WriteString(fmt.Sprintf("%s :%s: ", name, m.IconName(name)))
		}
	}
	return buf.String()
}

func (m *NotificationItem) Footer() (footer string) {
	createdAt := m.pr.CreatedAt.Local().Format(time.RFC1123)
	updatedAt := m.pr.UpdatedAt.Local().Format(time.RFC1123)

	if createdAt == updatedAt {
		footer = fmt.Sprintf("Created at %s.", createdAt)
	} else {
		footer = fmt.Sprintf("Created at %s. Updated at %s.", createdAt, updatedAt)
	}
	return
}

func (m *NotificationItem) MakeAttachment() slack.Attachment {
	return slack.Attachment{
		Color:     m.Color(),
		Title:     m.Title(),
		TitleLink: *m.pr.HTMLURL,
		Text:      m.ReactionTable(),
		Footer:    m.Footer(),
	}
}
