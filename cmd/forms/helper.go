package forms

import (
	"context"
	"os"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v70/github"
	"golang.org/x/term"
)

var (
	sO       string
	tO       string
	sR       string
	tR       string
	orgList  []string
	repoList []string
	err      error
	confirm  bool
	migrate  bool
	token    string
	client   *github.Client
)

var ctx = context.Background()
var success = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
var failure = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
var sp = spinner.New().Context(ctx).Title("")

func TerminalHeightHelper() int {
	// Get terminal height dynamically
	_, terminalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		terminalHeight = 20 // fallback height if unable to get terminal size
	}
	return terminalHeight
}
