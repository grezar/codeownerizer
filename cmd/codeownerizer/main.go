package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v44/github"
	"github.com/grezar/codeownerizer"
	"github.com/hmarr/codeowners"
	"golang.org/x/oauth2"
)

var (
	org  string
	repo string
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	ruleset, err := codeowners.LoadFileFromStandardLocation()
	if err != nil {
		return err
	}

	var owners []codeowners.Owner
	for _, rule := range ruleset {
		owners = append(owners, rule.Owners...)
	}

	if os.Getenv("CI") == "true" && os.Getenv("GITHUB_ACTION") != "" {
		// GITHUB_REPOSITORY is the owner and repository name. For example, octocat/Hello-World.
		githubRepository := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
		org = githubRepository[0]
		repo = githubRepository[1]
	} else {
		flag.StringVar(&org, "org", "", "GitHub organization")
		flag.StringVar(&repo, "repo", "", "GitHub repository")
		flag.Parse()
	}

	return codeownerizer.AddUngrantedOwners(ctx, client, org, repo, owners)
}
