package codeownerizer

import "github.com/google/go-github/v44/github"

func hasTeamOwnerSufficientPermission(teams []*github.Team, name string) bool {
	for _, team := range teams {
		if (stringify(team.Name) == name) && !team.Permissions[pushPermission] {
			return false
		}
	}
	return true
}

func hasUserOwnerSufficientPermission(collaborators []*github.User, name string) bool {
	for _, collaborator := range collaborators {
		if (stringify(collaborator.Name) == name) && !collaborator.Permissions[pushPermission] {
			return false
		}
	}
	return true
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
