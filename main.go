package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

const (
	helptext = `
ght is a tool for interacting with github repos from the command line.

Usage:

	ght <command> [arguments]

The commands are:

	repos       		list the repositories
	help        		show this help
	help [command]	show help for command

Configuration:

	Requires a GitHub Personal API token (https://blog.github.com/2013-05-16-personal-api-tokens/)
	in file ~/.ght with rights to access the repositories in question.
`
)

func init() {
	// Remove the Date/Time from log messages
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	// Log to stdout
	log.SetOutput(os.Stdout)

}

func main() {

	var err error
	// Subcommands
	reposCommand := flag.NewFlagSet("repos", flag.ExitOnError)
	orgPtr := reposCommand.String("o", "", "Specify the GitHub organisation")
	userPtr := reposCommand.String("u", "", "Specify the GitHub user")
	repoCommand := flag.NewFlagSet("repo", flag.ExitOnError)
	numReleasesPtr := reposCommand.Int("n", 20, "Specify the maximum number of releases to display")

	// Verify that a subcommand has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if len(os.Args) < 2 {
		log.Println(helptext)
		os.Exit(1)
	}
	// Switch on the subcommand
	// Parse the flags for appropriate FlagSet
	// FlagSet.Parse() requires a set of arguments to parse as input
	// os.Args[2:] will be all arguments starting after the subcommand at os.Args[1]
	switch os.Args[1] {
	case "help":
		if len(os.Args) > 2 {
			switch os.Args[2] {
			case "repos":
				err = doListRepos(reposCommand, orgPtr, userPtr, true)

			case "repo":
				err = doRepo(repoCommand, *numReleasesPtr, true)

			default:
				log.Printf("Help unknown command '%s'", os.Args[2])
				log.Println(helptext)
				os.Exit(1)

			}
		} else {
			log.Println(helptext)
		}
	case "repos":
		reposCommand.Parse(os.Args[2:])
		err = doListRepos(reposCommand, orgPtr, userPtr, false)

	case "repo":
		repoCommand.Parse(os.Args[2:])
		err = doRepo(repoCommand, *numReleasesPtr, false)

	default:
		log.Printf("Unknown command '%s'", os.Args[1])
		log.Println(helptext)
		os.Exit(1)
	}

	// Exit non zero on error
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func getClient() (*github.Client, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(usr.HomeDir, ".ght")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("ght. Missing '%s' file. This should contain your GitHub Personal API token. See https://blog.github.com/2013-05-16-personal-api-tokens/\n", path)
	}

	token, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ght. Error reading file '%s', error: %s\n", path, err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(string(token))},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	if err != nil {
		return nil, fmt.Errorf("ght. Error creating client: %s\n", err)
	}

	return client, nil
}

func doListRepos(flags *flag.FlagSet, org *string, user *string, displayHelp bool) error {

	helptext := `
ght repos 		List github repositories for an organisation or user

Usage:

	mdd repos [arguments]

The arguments are:
`

	// Asked for help?
	if displayHelp {
		log.Println(helptext)
		flags.PrintDefaults()
		return nil
	}

	// FlagSet.Parse() will evaluate to false if no flags were parsed
	if !flags.Parsed() {
		return fmt.Errorf("Error parsing arguments")
	}

	if (*org == "") == (*user == "") {
		return fmt.Errorf("Invalid arguments. Provide one of '-o organisation' or '-u user'")
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	var allRepos []*github.Repository
	if *org != "" {
		allRepos, err = listReposByOrg(client, *org)
	} else if *user != "" {
		allRepos, err = listReposByUser(client, *user)
	}
	if err != nil {
		return err
	}
	for _, r := range allRepos {
		log.Printf("%s\n", *r.FullName)
	}
	return nil
}

func doRepo(flags *flag.FlagSet, maxReleases int, displayHelp bool) error {

	helptext := `
ght repo 		Summarise a given repository

Usage:

	mdd repo owner repo

`

	// Asked for help?
	if displayHelp {
		log.Println(helptext)
		flags.PrintDefaults()
		return nil
	}

	// FlagSet.Parse() will evaluate to false if no flags were parsed
	if !flags.Parsed() {
		return fmt.Errorf("Error parsing arguments")
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	owner := os.Args[2]
	reponame := os.Args[3]

	ctx := context.Background()
	repo, _, err := client.Repositories.Get(ctx, owner, reponame)
	if err != nil {
		return err
	}

	log.Printf("FullName           %s\n", *repo.FullName)
	log.Printf("DefaultBranch      %s\n", *repo.DefaultBranch)

	protection, _, err := client.Repositories.GetBranchProtection(ctx, owner, reponame, *repo.DefaultBranch)
	if protection != nil {
		prReviews := protection.GetRequiredPullRequestReviews()
		log.Printf("BranchProtection/RequireCodeOwnerReviews       %t\n", prReviews.RequireCodeOwnerReviews)
		log.Printf("BranchProtection/RequiredApprovingReviewCount  %d\n", prReviews.RequiredApprovingReviewCount)
	} else {
		log.Printf("BranchProtection   None\n")
	}

	releases, err := listReleases(client, owner, reponame, maxReleases)
	if err != nil {
		return err
	}
	log.Printf("\nReleases:\n---------\n")
	tmpl := "%-19s %-12s %-18s %-40s\n"
	log.Printf(tmpl, "Published", "Tag", "Author", "Name")
	for i, release := range releases {
		if i >= maxReleases {
			break
		}
		log.Printf(tmpl, formatDate(release.PublishedAt), *release.TagName, *release.Author.Login, release.GetName())
	}

	return nil
}

func formatDate(t *github.Timestamp) string {
	return t.In(time.Local).Format("2006-01-02 15:04:05")
}

func listReleases(client *github.Client, owner, repo string, max int) ([]*github.RepositoryRelease, error) {
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 100}
	// get all pages of results
	var allReleases []*github.RepositoryRelease
	for {
		releases, resp, err := client.Repositories.ListReleases(ctx, owner, repo, opt)
		if err != nil {
			return allReleases, err
		}
		allReleases = append(allReleases, releases...)
		if resp.NextPage == 0 {
			break
		}
		// Break after retrieved max, note sreturned size might be a bit larger than max, because
		// we retrieve a page at a time
		if max > 0 && len(allReleases) > max {
			break
		}
		opt.Page = resp.NextPage
	}
	return allReleases, nil
}

func listReposByUser(client *github.Client, user string) ([]*github.Repository, error) {
	ctx := context.Background()
	opt := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 10},
	}
	// get all pages of results
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List(ctx, user, opt)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}

func listReposByOrg(client *github.Client, org string) ([]*github.Repository, error) {
	ctx := context.Background()
	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 10},
	}
	// get all pages of results
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}
