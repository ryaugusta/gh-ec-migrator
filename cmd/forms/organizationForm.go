package forms

import (
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/google/go-github/v70/github"
)

func OrganizationSelection() {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Source organization").
				SuggestionsFunc(func() []string {
					orgs, _, err := client.Organizations.List(ctx, "", nil)
					if err != nil {
						if err == huh.ErrUserAborted {
							log.Fatal("User aborted the program")
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
					_, _, err := client.Organizations.Get(ctx, tO)
					if err != nil {
						if err.(*github.ErrorResponse).Response.StatusCode == 404 {
							return fmt.Errorf("organization: %v does not exist", tO)
						}
					}

					if sO == tO {
						return fmt.Errorf("the source and target organizations have to be different")
					}

					return nil
				}),
		),
	).WithTheme(huh.ThemeDracula()).WithWidth(35).WithLayout(huh.LayoutColumns(2)).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("User aborted the program")
		}
		log.Fatalf("[ERROR]: %v", err)
	}
	RepositorySelection()
}
