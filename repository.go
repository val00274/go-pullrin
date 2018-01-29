package pullrin

import (
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type Repository struct {
	client *github.Client
	owner  string
	repo   string
}

func OpenRepository(ctx context.Context, token, owner, repo string) Repository {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return Repository{
		client: github.NewClient(tc),
		owner:  owner,
		repo:   repo,
	}
}

func (r *Repository) FetchOpenPullRequests(ctx context.Context) ([]*github.PullRequest, error) {
	result, _, err := r.client.PullRequests.List(ctx, r.owner, r.repo, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) FetchPullRequestReactions(ctx context.Context, i *github.PullRequest) ([]*github.Reaction, error) {
	reactions, _, err := r.client.Reactions.ListIssueReactions(ctx, r.owner, r.repo, *i.Number, nil)
	if err != nil {
		return nil, err
	}

	return reactions, nil
}

func (r *Repository) FetchPullRequestReactionNameToContent(ctx context.Context, i *github.PullRequest) (map[string]string, error) {
	reactions, err := r.FetchPullRequestReactions(ctx, i)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for _, reaction := range reactions {
		result[*reaction.User.Login] = *reaction.Content
	}
	return result, nil
}

func (r *Repository) FetchCommentCount(ctx context.Context, i *github.PullRequest) (c int) {
	if comments, _, err := r.client.Issues.ListComments(ctx, r.owner, r.repo, *i.Number, nil); err == nil {
		c += len(comments)
	} else {
		panic(err)
	}

	if comments, _, err := r.client.PullRequests.ListComments(ctx, r.owner, r.repo, *i.Number, nil); err == nil {
		c += len(comments)
	} else {
		panic(err)
	}

	return
}
