package forms

import (
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/google/go-github/v70/github"
)

func PostMigration() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Start Post-Migration Activities?").
				Affirmative("Continue").
				Negative("No").
				Value(&confirm).
				WithAccessible(true),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR] %v", err)
	}

	switch confirm {
	case true:
		selectActivity()
	case false:
		log.Fatal("Not continuing with post-migration activities. Goodbye!")
	}
}

func selectActivity() {
	var activity string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Activity").
				Options(
					huh.NewOption("Add Collaborators", "collaborators"),
					huh.NewOption("Reclaim Mannequins", "mannequins"),
				).
				Value(&activity),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR] %v", err)
	}

	switch activity {
	case "collaborators":
		addCollaboratorsForm()
	case "mannequins":
		reclaimMannequins()
	}
}

func addCollaboratorsForm() {
	log.Println("[DEBUG] Add collaborators")
	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Who would you like to add?").
				Options(
					huh.NewOption("Teams", "teams"),
					huh.NewOption("Users", "users"),
				).Value(&choice),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR] %v", err)
	}

	switch choice {
	case "teams":
		var teamList []string
		teamList = []string{}
		opts := &github.ListOptions{PerPage: 50}
		for {
			teams, resp, err := client.Teams.ListTeams(ctx, tO, opts)
			if err != nil {
				log.Fatalf("[ERROR] Error fetching teams: %v", err)
			}

			for _, t := range teams {
				teamList = append(teamList, t.GetSlug())
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		var choices []string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select The Team(s)").
					Description("You can select multiple teams here").
					Options(huh.NewOptions(teamList...)...).
					Value(&choices),
			),
		).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				log.Fatal("[WARN] User aborted the program")
			}
			log.Fatalf("[ERROR] %v", err)
		}
		addTeamCollaborators(choices)

	case "users":
		var userList []string
		userList = []string{}
		opts := &github.ListMembersOptions{
			ListOptions: github.ListOptions{PerPage: 50},
		}
		for {
			users, resp, err := client.Organizations.ListMembers(ctx, tO, opts)
			if err != nil {
				log.Fatalf("[ERROR] Error fetching teams: %v", err)
			}

			for _, u := range users {
				userList = append(userList, u.GetLogin())
			}

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		var choices []string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select The User(s)").
					Description("You can select multiple users here").
					Options(huh.NewOptions(userList...)...).
					Value(&choices),
			),
		).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				log.Fatal("[WARN] User aborted the program")
			}
			log.Fatalf("[ERROR] %v", err)
		}

		addUserCollaborators(choices)
	}
}

func addTeamCollaborators(choices []string) {
	var perm string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Permission").
				Description("What permission should the collaborator(s) have?").
				Options(
					huh.NewOption("Read", "pull"),
					huh.NewOption("Write", "push"),
					huh.NewOption("Admin", "admin"),
					huh.NewOption("Maintain", "maintain"),
					huh.NewOption("Triage", "triage"),
				).Value(&perm),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR] %v", err)
	}

	permission := &github.TeamAddTeamRepoOptions{Permission: perm}
	for _, c := range choices {
		log.Printf("[DEBUG] Adding Collaborator: %v", c)
		_, err := client.Teams.AddTeamRepoBySlug(ctx, tO, c, tO, tR, permission)
		log.Printf("[INFO] Granted \"%v\", %v permission", c, permission)
		if err != nil {
			log.Printf("[ERROR] problem adding collaborator: %v", err)
		}
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add more collaborators?").
				Description("Give others access to the repository").
				Affirmative("Yes").
				Negative("No").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR] %v", err)
	}

	switch confirm {
	case true:
		addCollaboratorsForm()
	case false:
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Reclaim Mannequins?").
					Description("Start the reclaim mannequin process").
					Affirmative("Yes").
					Negative("No").
					Value(&confirm),
			),
		).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				log.Fatal("[WARN] User aborted the program")
			}
			log.Fatalf("[ERROR] %v", err)
		}

		switch confirm {
		case true:
			// run reclaim mannequin process
			reclaimMannequins()
		case false:
			log.Fatalf("Nothing left to do. Goodbye!")
		}
	}
}

func addUserCollaborators(choices []string) {
	var perm string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Permission").
				Description("What permission should the collaborator(s) have?").
				Options(
					huh.NewOption("Read", "pull"),
					huh.NewOption("Write", "push"),
					huh.NewOption("Admin", "admin"),
					huh.NewOption("Maintain", "maintain"),
					huh.NewOption("Triage", "triage"),
				).Value(&perm),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR] %v", err)
	}

	permission := &github.RepositoryAddCollaboratorOptions{Permission: perm}
	tR = "test"
	for _, c := range choices {
		log.Printf("[DEBUG] Adding Collaborator: %v", c)
		_, _, err := client.Repositories.AddCollaborator(ctx, tO, tR, c, permission)
		log.Printf("[INFO] Granted \"%v\", %v permission", c, perm)
		if err != nil {
			log.Printf("[ERROR] problem adding collaborator: %v", err)
		}
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add more collaborators?").
				Description("Give others access to the repository").
				Affirmative("Yes").
				Negative("No").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("[WARN] User aborted the program")
		}
		log.Fatalf("[ERROR] %v", err)
	}

	switch confirm {
	case true:
		addCollaboratorsForm()
	case false:
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Reclaim Mannequins?").
					Description("Start the reclaim mannequin process").
					Affirmative("Yes").
					Negative("No").
					Value(&confirm),
			),
		).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				log.Fatal("[WARN] User aborted the program")
			}
			log.Fatalf("[ERROR] %v", err)
		}

		switch confirm {
		case true:
			// run reclaim mannequin process
			reclaimMannequins()
		case false:
			log.Fatalf("Nothing left to do. Goodbye!")
		}
	}
}

func reclaimMannequins() {
	fmt.Println("Reclaim Mannequins currently in development")
	fmt.Println("Use the GEI to reclaim mannequins here: https://docs.github.com/en/migrations/using-github-enterprise-importer/completing-your-migration-with-github-enterprise-importer/reclaiming-mannequins-for-github-enterprise-importer#reclaiming-mannequins-with-the-gei-extension")
}
