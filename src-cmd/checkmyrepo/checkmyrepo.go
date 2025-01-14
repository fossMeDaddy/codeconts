package checkmyrepoCmd

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
)

var (
	username string
	password string
)

var checkmyrepoCmd = &cobra.Command{
	Use:   "check [repository URL]",
	Short: "command to check the code contribution of all developers",
	Args:  cobra.MaximumNArgs(1),
	Run:   cmdRun,
	Example: `  checkmyrepo check                               # Check local repository
  checkmyrepo check https://github.com/user/repo     # Check remote repository
  checkmyrepo check https://gitlab.com/group/repo     # Check GitLab repository
  checkmyrepo check https://bitbucket.org/workspace/repo  # Check Bitbucket repository
  checkmyrepo check user/repo                        # Check remote repository`,
}

func getRepo(repoURL string) (*git.Repository, string, error) {
	var auth *http.BasicAuth
	if username != "" && password != "" {
		auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}
	if repoURL == "" {
		repo, err := git.PlainOpen(".")
		return repo, "", err
	}
	if !strings.Contains(repoURL, "://") {
		parts := strings.Split(repoURL, "/")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid format for repository name. Use owner/repo or full URL")
		}
		repoURL = fmt.Sprintf("https://github.com/%s/%s", parts[0], parts[1])
	}
	tempDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create a temp directory:%v", err)
	}
	fmt.Printf("Cloning repository %s...\n", repoURL)
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("failed to clone repository: %v", err)
	}
	return repo, tempDir, nil
}

func cmdRun(cmd *cobra.Command, args []string) {
	var repoURL string
	if len(args) > 0 {
		repoURL = args[0]
	}
	repo, tempDir, err := getRepo(repoURL)
	if err != nil {
		cmd.PrintErrf("Error:%v", err)
		return
	}
	if tempDir != "" {
		defer os.RemoveAll(tempDir)
	}
	ref, err := repo.Head()
	if err != nil {
		log.Fatalf("Failed to get HEAD reference: %v", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Fatalf("Failed to get commit history: %v", err)
	}
	authorChanges := make(map[string]int)
	var totalChanges int
	err = commitIter.ForEach(func(c *object.Commit) error {
		// fmt.Printf("Commit: %s\nAuthor: %s <%s>\nDate: %s\nMessage: %s\n\n",
		// 	c.Hash.String(),
		// 	c.Author.Name,
		// 	c.Author.Email,
		// 	c.Author.When,
		// 	c.Message,
		// )

		stats, err := c.Stats()
		if err != nil {
			log.Fatalf("Failed to get Stats: %v", err)
		}
		changes := 0
		for _, stat := range stats {
			changes += stat.Addition
		}
		author := fmt.Sprintf("%s <%s>", c.Author.Name, c.Author.Email)
		authorChanges[author] += changes
		totalChanges += changes

		return nil
	})
	if err != nil {
		log.Fatalf("Error while iterating commits: %v", err)
	}

	type AuthorStat struct {
		Author string
		Equity float64
	}
	var stats []AuthorStat
	for author, changes := range authorChanges {
		equity := (float64(changes) / float64(totalChanges)) * 100
		stats = append(stats, AuthorStat{Author: author, Equity: equity})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Equity > stats[j].Equity
	})
	fmt.Println("Developers of this Repo:")
	for _, stat := range stats {
		redValue := uint8((1 - (stat.Equity / 100)) * 255)
		greenValue := uint8((stat.Equity / 100) * 255)
		// equityColor := color.New().Add(color.Attribute(38)).Add(color.Attribute(48)).
		// 	Add(color.Attribute(5)).
		// 	Add(color.Attribute(16 + int(redValue/36)*36 + int(greenValue/36)*6)).
		// 	Sprint(fmt.Sprintf("%.2f%%", stat.Equity))
		equityColor := color.RGB(int(redValue), int(greenValue), 0).Sprintf(fmt.Sprintf("%.2f%%%%", stat.Equity))
		authorColor := color.New(color.FgMagenta).Sprint(stat.Author)

		fmt.Printf("    Author: %s owns Code Equity: %s\n", authorColor, equityColor)
	}

}

func Init() *cobra.Command {
	checkmyrepoCmd.Flags().StringVarP(&username, "username", "u", "", "Username for private repository")
	checkmyrepoCmd.Flags().StringVarP(&password, "password", "p", "", "Password or token for private repository")
	return checkmyrepoCmd
}
