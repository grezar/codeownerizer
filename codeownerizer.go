package codeownerizer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v44/github"
	"github.com/hmarr/codeowners"
)

const pushPermission = "push"

func AddUngrantedOwners(ctx context.Context, api *github.Client, org string, repo string, owners []codeowners.Owner) error {
	teams, _, err := api.Repositories.ListTeams(ctx, org, repo, nil)
	if err != nil {
		return err
	}

	collaborators, _, err := api.Repositories.ListCollaborators(ctx, org, repo, nil)
	if err != nil {
		return err
	}

	for _, owner := range owners {
		switch owner.Type {
		case codeowners.TeamOwner:
			teamOwnerName := strings.Split(owner.String(), "/")[1]

			if !hasTeamOwnerSufficientPermission(teams, teamOwnerName) || !containsTeamOwner(teams, teamOwnerName) {
				resp, err := api.Teams.AddTeamRepoBySlug(ctx, org, teamOwnerName, org, repo, &github.TeamAddTeamRepoOptions{
					Permission: pushPermission,
				})
				if err != nil {
					return err
				}
				if err = github.CheckResponse(resp.Response); err != nil {
					return err
				}
			}
		case codeowners.UsernameOwner:
			userOwnerName := strings.TrimPrefix(owner.String(), "@")

			if !hasUserOwnerSufficientPermission(collaborators, userOwnerName) || !containsUserOwner(collaborators, userOwnerName) {
				_, resp, err := api.Repositories.AddCollaborator(ctx, org, repo, userOwnerName, &github.RepositoryAddCollaboratorOptions{
					Permission: pushPermission,
				})
				if err != nil {
					return err
				}
				if err = github.CheckResponse(resp.Response); err != nil {
					return err
				}
			}
		case codeowners.EmailOwner:
			emailOwnerEmail := owner.String()
			userSearchResult, resp, err := api.Search.Users(ctx, fmt.Sprintf("%s in:email", emailOwnerEmail), nil)
			if err != nil {
				return err
			}
			if err = github.CheckResponse(resp.Response); err != nil {
				return err
			}
			if len(userSearchResult.Users) > 1 {
				return errors.New("multiple users who has %s in email was found")
			}

			emailOwnerUsername := stringify(userSearchResult.Users[0].Name)

			if !hasUserOwnerSufficientPermission(collaborators, emailOwnerUsername) || !containsUserOwner(collaborators, emailOwnerUsername) {
				_, resp, err := api.Repositories.AddCollaborator(ctx, org, repo, emailOwnerUsername, &github.RepositoryAddCollaboratorOptions{
					Permission: pushPermission,
				})
				if err != nil {
					return err
				}
				if err = github.CheckResponse(resp.Response); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unknown owner type: %s\n", owner.Type)
		}
	}

	return nil
}
