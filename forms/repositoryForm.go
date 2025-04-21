package forms

import (
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/google/go-github/v70/github"
)

func RepositorySelection() {
	loadRepos := func() {
		sp.Title("Loading Repositories...")
		repoList = []string{}
		opts := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 50},
		}

		for {
			repos, resp, err := client.Repositories.ListByOrg(ctx, sO, opts)
			if err != nil {
				log.Fatalf("Error fetching repositories: %v", err)
			}

			for _, r := range repos {
				repoList = append(repoList, r.GetName())
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}

	sp.Action(loadRepos).Run()

	repositoryOpts := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a repository").
				Description("The repository you want to migrate").
				Options(huh.NewOptions(repoList...)...).
				Value(&sR),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Enter the target repo name").
				Description("This is the name the migrated repository will be").
				Validate(func(s string) error {
					_, _, err := client.Repositories.Get(ctx, tO, tR)
					if err != nil {
						// if repository doesn't exist, it's available!
						if _, ok := err.(*github.ErrorResponse); ok && err.(*github.ErrorResponse).Response.StatusCode == 404 {
							return nil
						}

						fmt.Println(err)
						log.Fatal("quitting program")

						// print other errors not handled
						return fmt.Errorf("error checking repository availability: %v", err)
					}
					// If no error, the repository exists
					return fmt.Errorf("repository name '%v' is already taken in target organization '%v'. Please choose a different name", tR, tO)
				}).Value(&tR),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := repositoryOpts.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("User aborted the program")
		}
		log.Fatalf("[ERROR]: %v", err)
	}

	Migration()
}
