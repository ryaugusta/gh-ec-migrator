package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v70/github"
	"github.com/joho/godotenv"
)

var (
	token    string
	sO       string
	tO       string
	sR       string
	tR       string
	client   *github.Client
	orgList  []string
	repoList []string
	migrate  bool
	err      error
)

var ctx = context.Background()
var success = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
var failure = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
var sp = spinner.New().Context(ctx).Title("")

func main() {

	var authMethod string
	authForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Authentication").
				Description("Select your authentication method").
				Options(
					huh.NewOption("Personal Access Token", "pat"),
					huh.NewOption("Environment Variable", "env"),
				).Value(&authMethod),
		),
	).WithTheme(huh.ThemeDracula())
	authForm.Run()

	// show form based on selection
	switch authMethod {
	case "pat":
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Authentication").
					Description("Enter your PAT").
					Value(&token).
					Validate(func(s string) error {
						client = github.NewClient(nil).WithAuthToken(token)
						if token == "" {
							return fmt.Errorf("the token cannot be empty")
						}
						return nil
					}),
			),
		).WithWidth(80)
		form.Run()
	case "env":
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		token = os.Getenv("GITHUB_TOKEN")
		client = github.NewClient(nil).WithAuthToken(token)
		log.Output(0, "token set")
	}

	orgForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Source organization").
				SuggestionsFunc(func() []string {
					orgs, _, err := client.Organizations.List(ctx, "", nil)
					if err != nil {
						if err == huh.ErrUserAborted {
							log.Fatal("user aborted")
						}

						fmt.Println(err)
						log.Fatal("quitting program")
					}

					for _, org := range orgs {
						orgList = append(orgList, org.GetLogin())
					}
					return orgList
				}, &sO).
				Value(&sO),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Target organization").
				SuggestionsFunc(func() []string {
					return orgList
				}, &tO).
				Value(&tO).
				Validate(func(s string) error {
					if sO == tO {
						return fmt.Errorf("the source and target organizations have to be different")
					}
					// _, _, err := client.Organizations.Get(ctx, tO)
					// if err.(*github.ErrorResponse).Response.StatusCode == 404 {
					// 	return fmt.Errorf("%v organization does not exist", tO)
					// }

					return nil
				}),
		),
	).WithTheme(huh.ThemeDracula()).WithWidth(35).WithLayout(huh.LayoutColumns(2))

	orgForm.Run()

	repoList = []string{}
	loadRepos := func() {
		sp.Title("Loading Repositories...")
		repos, _, _ := client.Repositories.ListByOrg(ctx, sO, &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 30},
		})

		for _, r := range repos {
			repoList = append(repoList, r.GetName())
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

						if err == huh.ErrUserAborted {
							log.Fatal("user aborted")
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
	).WithTheme(huh.ThemeDracula()).WithWidth(80)

	repositoryOpts.Run()

	confirm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("done").
				Title("Ready to Migrate?").
				Validate(func(v bool) error {
					if !v {
						return fmt.Errorf("finish up, or ctrl+c to quit")
					}
					return nil
				}).
				Affirmative("Migrate!").
				Negative("Not ready yet.").
				Value(&migrate).
				WithAccessible(true),
		),
	).WithTheme(huh.ThemeDracula()).WithWidth(50)

	confirm.Run()

	// now we are out of the forms:
	if !migrate {
		if err == huh.ErrUserAborted {
			log.Fatal("user aborted")
		}
		log.Fatalf("Not starting the migration:  migrate = %v", migrate)
	}

	sp = spinner.New().Context(ctx).Title("Starting Migration...")
	upg := func() {
		sp.Title("Upgrading extension...")
		out, err := exec.Command("gh", "extension", "upgrade", "gei").Output()
		if err != nil {
			log.Fatalf("Error upgrading extension: %v\n Output: %v", err, out)
			if err == huh.ErrUserAborted {
				log.Fatal("user aborted")
			}
		}
	}

	migration := func() {
		sp.Title("Migration in progress...")
		// write log to:
		if err := os.Truncate("m.log", 0); err != nil {
			log.Printf("Failed to truncate: %v", err)
		}
		f, err := os.OpenFile("m.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777) // create, append, update perms to file

		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		mw := io.MultiWriter(f, os.Stdout)

		cmd := exec.Command("gh", "gei", "migrate-repo",
			// "--help")
			"--github-source-org="+sO,
			"--github-target-org="+tO,
			"--source-repo="+sR,
			"--target-repo="+tR,
			"--github-target-pat="+token)
		cmd.Stderr = mw
		cmd.Stdout = mw
		err = cmd.Run()
		if err != nil {
			log.Output(0, failure.Render("Migration failed."))
		}
		readLog, _ := os.ReadFile("m.log")
		if strings.Contains(string(readLog), "fail") {
			log.Output(0, failure.Render("Migration failed."))
		}
		log.Output(0, success.Render("Migration successful!"))
	}

	sp.Action(upg).Run()
	sp.Action(migration).Run()
}
