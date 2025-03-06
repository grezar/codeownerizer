package codeownerizer

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v69/github"
	"github.com/hmarr/codeowners"

	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestAddUngrantedOwners(t *testing.T) {
	org := "org"
	repo := "repo"

	ruleset, err := codeowners.LoadFile("testdata/CODEOWNERS")
	if err != nil {
		t.Error(err)
	}
	var owners []codeowners.Owner
	for _, rule := range ruleset {
		owners = append(owners, rule.Owners...)
	}

	ctx := context.Background()

	var octocatsTeamAddedToRepo bool
	putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsTeam := mock.EndpointPattern{
		Pattern: fmt.Sprintf("/orgs/%s/teams/%s/repos/%s/%s", org, "octocats", org, repo),
		Method:  "PUT",
	}

	var octocatsAdminsTeamAddedToRepo bool
	putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsAdminsTeam := mock.EndpointPattern{
		Pattern: fmt.Sprintf("/orgs/%s/teams/%s/repos/%s/%s", org, "octocats-admins", org, repo),
		Method:  "PUT",
	}

	var missingInsufficientPermissionUserOwnerAddedToRepo bool
	putReposCollaboratorsByOwnerByRepoWithDoctocat := mock.EndpointPattern{
		Pattern: fmt.Sprintf("/repos/%s/%s/collaborators/%s", org, repo, "doctocat"),
		Method:  "PUT",
	}

	var missingUserOwnerAddedToRepo bool
	putReposCollaboratorsByOwnerByRepoWithOctocat := mock.EndpointPattern{
		Pattern: fmt.Sprintf("/repos/%s/%s/collaborators/%s", org, repo, "octocat"),
		Method:  "PUT",
	}

	var missingEmailOwnerAddedToRepo bool
	putReposCollaboratorsByOwnerByRepoWithEmailOwner := mock.EndpointPattern{
		Pattern: fmt.Sprintf("/repos/%s/%s/collaborators/%s", org, repo, "email-owner"),
		Method:  "PUT",
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposTeamsByOwnerByRepo,
			[]github.Team{
				{
					Name: github.Ptr("octocats"),
					Slug: github.Ptr("octocats"),
				},
				{
					Name: github.Ptr("Octocats Admins"),
					Slug: github.Ptr("octocats-admins"),
				},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposCollaboratorsByOwnerByRepo,
			[]github.User{
				{
					Login: github.Ptr("global-owner1"),
					Permissions: map[string]bool{
						"push": true,
					},
				},
				{
					Login: github.Ptr("global-owner2"),
					Permissions: map[string]bool{
						"push": true,
					},
				},
				{
					Login: github.Ptr("js-owner"),
					Permissions: map[string]bool{
						"push": true,
					},
				},
				// Assume doctocat don't have sufficient permission as a CODEOWNER
				{
					Login: github.Ptr("doctocat"),
					Permissions: map[string]bool{
						"push": false,
					},
				},
				// Assume octocat and docs@example.com is listed in CODEOWNERS but not granted to push.
				// These owners should be added to the repository with the right permission.
				// {
				// 	Name: github.String("octocat"),
				// },
				// {
				// 	Email: github.String("docs@example.com"),
				// },
			},
		),
		mock.WithRequestMatchHandler(
			putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsTeam,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				octocatsTeamAddedToRepo = true
			}),
		),
		mock.WithRequestMatchHandler(
			putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsAdminsTeam,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				octocatsAdminsTeamAddedToRepo = true
			}),
		),
		mock.WithRequestMatchHandler(
			putReposCollaboratorsByOwnerByRepoWithDoctocat,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				missingInsufficientPermissionUserOwnerAddedToRepo = true
			}),
		),
		mock.WithRequestMatchHandler(
			putReposCollaboratorsByOwnerByRepoWithOctocat,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				missingUserOwnerAddedToRepo = true
			}),
		),
		mock.WithRequestMatchHandler(
			mock.GetSearchUsers,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()["q"]
				if diff := cmp.Diff(q, []string{"docs@example.com in:email"}); diff != "" {
					t.Errorf("unexpected query value\n%s", diff)
				}

				_, _ = w.Write(mock.MustMarshal(github.UsersSearchResult{
					Users: []*github.User{
						{
							Name: github.Ptr("email-owner"),
						},
					},
				}))
			}),
		),
		mock.WithRequestMatchHandler(
			putReposCollaboratorsByOwnerByRepoWithEmailOwner,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				missingEmailOwnerAddedToRepo = true
			}),
		),
	)
	client := github.NewClient(mockedHTTPClient)

	err = AddUngrantedOwners(ctx, client, "org", "repo", owners)
	if err != nil {
		t.Error(err)
	}

	if !octocatsTeamAddedToRepo {
		t.Errorf(
			"expected %s %s to be called\n",
			putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsTeam.Method,
			putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsTeam.Pattern,
		)
	}

	if !octocatsAdminsTeamAddedToRepo {
		t.Errorf(
			"expected %s %s to be called\n",
			putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsAdminsTeam.Method,
			putOrgsTeamsReposByOrgByTeamSlugByOwnerByRepoWithOctocatsAdminsTeam.Pattern,
		)
	}

	if !missingInsufficientPermissionUserOwnerAddedToRepo {
		t.Errorf(
			"expected %s %s to be called\n",
			putReposCollaboratorsByOwnerByRepoWithDoctocat.Method,
			putReposCollaboratorsByOwnerByRepoWithDoctocat.Pattern,
		)
	}

	if !missingUserOwnerAddedToRepo {
		t.Errorf(
			"expected %s %s to be called\n",
			putReposCollaboratorsByOwnerByRepoWithOctocat.Method,
			putReposCollaboratorsByOwnerByRepoWithOctocat.Pattern,
		)
	}

	if !missingEmailOwnerAddedToRepo {
		t.Errorf(
			"expected %s %s to be called\n",
			putReposCollaboratorsByOwnerByRepoWithEmailOwner.Method,
			putReposCollaboratorsByOwnerByRepoWithEmailOwner.Pattern,
		)
	}
}
