package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/underwoo16/gh-diffstack/gh"
	"github.com/underwoo16/gh-diffstack/git"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current stack of diffs",
	Long: `Show the current stack of diffs.
Includes the following:

- The branch name associated with the diff
- The commit sha
- The commit message
- The pull request url (if it exists)`,
	Run: runShowCmd,
}

// TODO: cleanup
// TODO: add colors
func runShowCmd(cmd *cobra.Command, args []string) {
	gitService := git.NewGitService()
	out, err := gitService.LogFromMain()
	if err != nil {
		log.Fatal(err)
	}

	logString := string(out)
	logs := strings.FieldsFunc(logString, func(r rune) bool {
		return r == '\n'
	})

	stacks := []*diffStack{}
	for _, log := range logs {
		parts := strings.Fields(log)

		sha := strings.TrimSpace(parts[0])

		commitMessage := strings.Join(parts[1:], " ")

		// build diffstack
		diffStack := diffStack{
			sha:           sha,
			commitMessage: commitMessage,
			branchName:    gitService.BuildBranchNameFromCommit(sha),
		}
		stacks = append(stacks, &diffStack)
	}

	ghService := gh.NewGitHubService()

	// TODO: cache pull requests when created
	// only call api if no cache found
	pullRequests := ghService.GetPullRequests()

	for _, stack := range stacks {
		for _, pr := range pullRequests {
			if pr.HeadRefName == stack.branchName {
				stack.prUrl = pr.Url
			}
		}
	}

	sb := strings.Builder{}
	for idx, stack := range stacks {
		marker := circle
		if idx == 0 {
			marker = dot
		}

		sb.WriteString(fmt.Sprintf("%s %s (%s)\n", marker, stack.branchName, stack.sha))

		sb.WriteString(vertical)

		if stack.prUrl != "" {
			sb.WriteString(fmt.Sprintf("  - %s\n%s", stack.prUrl, vertical))
		}

		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("%s %s\n", trunk, "trunk"))

	fmt.Print(sb.String())

}

var vertical = "│"
var trunk = "┴"
var circle = "◌"
var dot = "●"

type diffStack struct {
	sha           string
	branchName    string
	commitMessage string
	prUrl         string
}