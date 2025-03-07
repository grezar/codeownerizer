package codeownerizer

import (
	"github.com/google/go-github/v69/github"
	"github.com/hmarr/codeowners"
)

func uniqueOwners(owners []codeowners.Owner) []codeowners.Owner {
	var unique []codeowners.Owner
	m := make(map[string]bool)
	for _, owner := range owners {
		v := owner.String()
		if !m[v] {
			m[v] = true
			unique = append(unique, owner)
		}
	}
	return unique
}

func hasTeamOwnerSufficientPermission(teams []*github.Team, owner string) bool {
	for _, team := range teams {
		if (stringify(team.Slug) == owner) && team.Permissions[pushPermission] {
			return true
		}
	}
	return false
}

func hasUserOwnerSufficientPermission(collaborators []*github.User, owner string) bool {
	for _, collaborator := range collaborators {
		if (stringify(collaborator.Login) == owner) && collaborator.Permissions[pushPermission] {
			return true
		}
	}
	return false
}

func containsTeamOwner(teams []*github.Team, owner string) bool {
	for _, team := range teams {
		if stringify(team.Slug) == owner {
			return true
		}
	}
	return false
}

func containsUserOwner(collaborators []*github.User, owner string) bool {
	for _, collaborator := range collaborators {
		if stringify(collaborator.Login) == owner {
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
