package codeownerizer

import (
	"context"
	"fmt"
	"strings"

	"log"

	"github.com/google/go-github/v69/github"
	"github.com/hmarr/codeowners"
)

const pushPermission = "push"

func AddUngrantedOwners(ctx context.Context, api *github.Client, org string, repo string, owners []codeowners.Owner) error {
	owners = uniqueOwners(owners)

	teams, err := ListTeams(ctx, api, org, repo)
	if err != nil {
		return err
	}

	collaborators, err := ListCollaborators(ctx, api, org, repo)
	if err != nil {
		return err
	}

	for _, owner := range owners {
		switch owner.Type {
		case codeowners.TeamOwner:
			teamOwnerName := strings.Split(owner.String(), "/")[1]

			// Grant a push permission to
			// - a team that is already have an access to the repository but does not have a push permission.
			// - a team that does not have an access to the repository.
			if !hasTeamOwnerSufficientPermission(teams, teamOwnerName) || !containsTeamOwner(teams, teamOwnerName) {
				resp, err := api.Teams.AddTeamRepoBySlug(ctx, org, teamOwnerName, org, repo, &github.TeamAddTeamRepoOptions{
					Permission: pushPermission,
				})
				if err != nil {
					log.Println(err.Error())
					continue
				}
				if err = github.CheckResponse(resp.Response); err != nil {
					log.Println(err.Error())
					continue
				}
				log.Printf("%s was added to the repo with the %s permission.\n", owner.String(), pushPermission)
			}
		case codeowners.UsernameOwner:
			userOwnerName := strings.TrimPrefix(owner.String(), "@")

			if !hasUserOwnerSufficientPermission(collaborators, userOwnerName) || !containsUserOwner(collaborators, userOwnerName) {
				_, resp, err := api.Repositories.AddCollaborator(ctx, org, repo, userOwnerName, &github.RepositoryAddCollaboratorOptions{
					Permission: pushPermission,
				})
				if err != nil {
					log.Println(err.Error())
					continue
				}
				if err = github.CheckResponse(resp.Response); err != nil {
					log.Println(err.Error())
					continue
				}
				log.Printf("%s was added to the repo with the %s permission.\n", owner.String(), pushPermission)
			}
		case codeowners.EmailOwner:
			emailOwnerEmail := owner.String()
			userSearchResult, resp, err := api.Search.Users(ctx, fmt.Sprintf("%s in:email", emailOwnerEmail), nil)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			if err = github.CheckResponse(resp.Response); err != nil {
				log.Println(err.Error())
				continue
			}
			if len(userSearchResult.Users) > 1 {
				log.Printf("multiple users who has %s in email was found\n", emailOwnerEmail)
				continue
			}

			emailOwnerUsername := stringify(userSearchResult.Users[0].Login)

			if !hasUserOwnerSufficientPermission(collaborators, emailOwnerUsername) || !containsUserOwner(collaborators, emailOwnerUsername) {
				_, resp, err := api.Repositories.AddCollaborator(ctx, org, repo, emailOwnerUsername, &github.RepositoryAddCollaboratorOptions{
					Permission: pushPermission,
				})
				if err != nil {
					log.Println(err.Error())
					continue
				}
				if err = github.CheckResponse(resp.Response); err != nil {
					log.Println(err.Error())
					continue
				}
				log.Printf("%s was added to the repo with the %s permission.\n", owner.String(), pushPermission)
			}
		default:
			log.Printf("unknown owner type: %s\n", owner.Type)
			continue
		}
	}

	return nil
}

func ListTeams(ctx context.Context, api *github.Client, org string, repo string) ([]*github.Team, error) {
	allTeams := []*github.Team{}
	opts := &github.ListOptions{PerPage: 100}
	for {
		teams, resp, err := api.Repositories.ListTeams(ctx, org, repo, opts)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allTeams, nil
}

func ListCollaborators(ctx context.Context, api *github.Client, org string, repo string) ([]*github.User, error) {
	allCollaborators := []*github.User{}
	opts := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		collaborators, resp, err := api.Repositories.ListCollaborators(ctx, org, repo, opts)
		if err != nil {
			return nil, err
		}
		allCollaborators = append(allCollaborators, collaborators...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allCollaborators, nil
}
