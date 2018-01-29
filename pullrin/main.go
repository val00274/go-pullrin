package main

import (
	"os"
	"strings"

	"github.com/nlopes/slack"
	"github.com/val00274/go-pullrin"
	"golang.org/x/net/context"
)

type env struct {
	name            string
	iconURL         string
	githubOwner     string
	githubRepo      string
	githubReviewers string
	githubToken     string
	slackAPIToken   string
	slackChannel    string
}

func getEnv() env {
	return env{
		name:            os.Getenv("PULLRIN_NAME"),
		iconURL:         os.Getenv("PULLRIN_ICON_URL"),
		githubOwner:     os.Getenv("PULLRIN_GITHUB_OWNER"),
		githubRepo:      os.Getenv("PULLRIN_GITHUB_REPO"),
		githubReviewers: os.Getenv("PULLRIN_GITHUB_REVIEWERS"),
		githubToken:     os.Getenv("PULLRIN_GITHUB_TOKEN"),
		slackAPIToken:   os.Getenv("PULLRIN_SLACK_API_TOKEN"),
		slackChannel:    os.Getenv("PULLRIN_SLACK_CHANNEL"),
	}
}

func (e *env) listReviewers() []string {
	return strings.Split(e.githubReviewers, ",")
}

func makeMessage(env *env) (string, []slack.Attachment, error) {
	ctx := context.Background()
	repo := pullrin.OpenRepository(ctx, env.githubToken, env.githubOwner, env.githubRepo)
	pulls, err := repo.FetchOpenPullRequests(ctx)
	if err != nil {
		return "", nil, err
	}
	if len(pulls) == 0 {
		return ":nico:", nil, nil
	}

	attachments := []slack.Attachment{}
	for _, pr := range pulls {
		reactions, err := repo.FetchPullRequestReactionNameToContent(ctx, pr)
		if err != nil {
			return "", nil, err
		}

		item := pullrin.NewNotificationItem(&repo, pr, reactions, env.listReviewers())

		attachments = append(attachments, item.MakeAttachment())
	}
	return "", attachments, nil
}

func main() {
	env := getEnv()

	ch := pullrin.OpenNotificationChannel(env.slackAPIToken, env.slackChannel)

	if message, attachments, err := makeMessage(&env); err == nil {
		ch.Post(env.name, env.iconURL, message, attachments)
	} else {
		ch.Post(env.name, env.iconURL, ":skull:", nil)
		panic(err)
	}
}
