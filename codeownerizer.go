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
			var insufficientPermission bool

			for _, team := range teams {
				if (stringify(team.Name) == teamOwnerName) && !team.Permissions[pushPermission] {
					insufficientPermission = true
				}
			}

			if !containsTeamOwner(teams, teamOwnerName) {
				insufficientPermission = true
			}

			if insufficientPermission {
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
			var insufficientPermission bool

			for _, collaborator := range collaborators {
				if (stringify(collaborator.Name) == userOwnerName) && !collaborator.Permissions[pushPermission] {
					insufficientPermission = true
				}
			}

			if !containsUserOwner(collaborators, userOwnerName) {
				insufficientPermission = true
			}

			if insufficientPermission {
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
			var insufficientPermission bool

			for _, collaborator := range collaborators {
				if (stringify(collaborator.Name) == emailOwnerUsername) && !collaborator.Permissions[pushPermission] {
					insufficientPermission = true
				}
			}

			if !containsUserOwner(collaborators, emailOwnerUsername) {
				insufficientPermission = true
			}

			if insufficientPermission {
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

func containsTeamOwner(s []*github.Team, e string) bool {
	for _, v := range s {
		if e == stringify(v.Name) {
			return true
		}
	}
	return false
}

func containsUserOwner(s []*github.User, e string) bool {
	for _, u := range s {
		if e == stringify(u.Name) {
			return true
		}
	}
	return false
}

func stringify(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
