package forms

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

func Migration() {
	form := huh.NewForm(
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
	).WithTheme(huh.ThemeDracula()).WithHeight(TerminalHeightHelper() - 5)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Fatal("User aborted the program")
		}
		log.Fatalf("[ERROR]: %v", err)
	}

	if !migrate {
		if err == huh.ErrUserAborted {
			log.Fatal("User aborted the program")
		}
		log.Fatalf("[ERROR]: %v", err)
	}

	sp = spinner.New().Context(ctx).Title("Starting Migration...")
	upg := func() {
		sp.Title("Upgrading extension...")
		out, err := exec.Command("gh", "extension", "upgrade", "gei").Output()
		if err != nil {
			log.Fatalf("Error upgrading extension: %v\n Output: %v", err, out)
			if err == huh.ErrUserAborted {
				log.Fatal("User aborted the program")
			}
		}
	}

	migration := func() {
		sp.Title("Migration in progress...")
		// write log to:
		if err := os.Truncate("m.log", 0); err != nil {
			log.Printf("Failed to truncate: %v\n", err)
			log.Println("Creating File...")
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
			log.Fatalf(failure.Render("Migration failed: %v"), err)
		}
		readLog, _ := os.ReadFile("m.log")
		if strings.Contains(string(readLog), "fail") {
			log.Fatalf(failure.Render("Migration failed: %v"), err)
		}
		log.Println(success.Render("Migration successful!"))

		// log the URL to the new repository
	}
	sp.Action(upg).Run()
	sp.Action(migration).Run()

	PostMigration() // start post migration
}
