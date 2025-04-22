package forms

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/google/go-github/v70/github"
	"github.com/joho/godotenv"
)

func AuthenticationForm() {
	var authMethod string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Authentication").
				Description("Select your authentication method").
				Options(
					huh.NewOption("Personal Access Token", "pat"),
					huh.NewOption("Environment Variable", "env"),
				).Value(&authMethod),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR]: %v", err)
	}

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
							return fmt.Errorf("[ERROR] the token cannot be empty")
						}
						return nil
					}),
			),
		).WithWidth(80)
		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				log.Fatal("User aborted the program")
			}
			log.Fatalf("[ERROR]: %v", err)
		}
	case "env":
		token = os.Getenv("GH_TOKEN")
		if token == "" {
			err := godotenv.Load()
			if err != nil {
				log.Fatal("[ERROR]: the ec-migrator expects 'GH_TOKEN' env variable to be set")
			}
		}

		token = os.Getenv("GH_TOKEN")
		client = github.NewClient(nil).WithAuthToken(token)

		OrganizationSelection()
	}
}
