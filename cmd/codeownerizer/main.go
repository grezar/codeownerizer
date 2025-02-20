package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v69/github"
	"github.com/grezar/codeownerizer"
	"github.com/hmarr/codeowners"
	"golang.org/x/oauth2"
)

var (
	// These variables are set in build step.
	Version  string
	Revision string

	version bool
	org     string
	repo    string
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flag.BoolVar(&version, "version", false, "Print version")
	flag.StringVar(&org, "org", "", "GitHub organization")
	flag.StringVar(&repo, "repo", "", "GitHub repository")
	flag.Parse()

	if version {
		fmt.Println("codeownerizer", Version, Revision)
		return nil
	}

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

	// Default values when it runs on GitHub Actions
	if (os.Getenv("CI") == "true") && (os.Getenv("GITHUB_ACTION") != "") {
		// GITHUB_REPOSITORY is the owner and repository name. For example, octocat/Hello-World.
		githubRepository := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")

		if org == "" {
			org = githubRepository[0]
		}

		if repo == "" {
			repo = githubRepository[1]
		}
	}

	return codeownerizer.AddUngrantedOwners(ctx, client, org, repo, owners)
}
