package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	_ "time"
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

type Repository struct {
	Name          string
	NameWithOwner string
}

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
	// repoCommand := flag.NewFlagSet("repo", flag.ExitOnError)
	// maxReleasesPtr := reposCommand.Int("maxr", 20, "Specify the maximum number of Releases to display")
	// maxTagsPtr := reposCommand.Int("maxt", 20, "Specify the maximum number of Tags to display")

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

			// case "repo":
			// 	err = doRepo(repoCommand, *maxReleasesPtr, *maxTagsPtr, true)

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

	// case "repo":
	// 	repoCommand.Parse(os.Args[2:])
	// 	err = doRepo(repoCommand, *maxReleasesPtr, *maxTagsPtr, false)

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

func getClient() (*githubv4.Client, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(usr.HomeDir, ".ght")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("ght. Missing '%s' file. This should contain your GitHub Personal API token. See https://blog.github.com/2013-05-16-personal-api-tokens/", path)
	}

	token, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ght. Error reading file '%s', error: %s", path, err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(string(token))},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := githubv4.NewClient(tc)
	if err != nil {
		return nil, fmt.Errorf("ght. Error creating client: %s", err)
	}

	return client, nil
}

func doListRepos(flags *flag.FlagSet, org *string, user *string, displayHelp bool) error {

	helptext := `
ght repos 		List github repositories for an organisation or user

Usage:

	mdd repos user-or-organisation [arguments]

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

	var allRepos []Repository
	if *org != "" {
		allRepos, err = listReposByOrg(client, *org)
	} else if *user != "" {
		allRepos, err = listReposByUser(client, *user)
	}
	if err != nil {
		return err
	}
	for _, r := range allRepos {
		log.Printf("%s\n", r.NameWithOwner)
	}
	return nil
}

// func doRepo(flags *flag.FlagSet, maxReleases, maxTags int, displayHelp bool) error {

// 	helptext := `
// ght repo 		Summarise a given repository

// Usage:

// 	mdd repo owner/repo

// `

// 	// Asked for help?
// 	if displayHelp {
// 		log.Println(helptext)
// 		flags.PrintDefaults()
// 		return nil
// 	}

// 	// FlagSet.Parse() will evaluate to false if no flags were parsed
// 	if !flags.Parsed() {
// 		return fmt.Errorf("Error parsing arguments")
// 	}

// 	client, err := getClient()
// 	if err != nil {
// 		return err
// 	}

// 	ownerRepo := strings.Split(os.Args[2], "/")
// 	if len(ownerRepo) != 2 {
// 		return fmt.Errorf("Error parsing %s as 'owner/repo'", os.Args[2])
// 	}
// 	owner := ownerRepo[0]
// 	reponame := ownerRepo[1]

// 	ctx := context.Background()
// 	repo, _, err := client.Repositories.Get(ctx, owner, reponame)
// 	if err != nil {
// 		return err
// 	}

// 	log.Printf("Full name :           %s\n", *repo.FullName)
// 	log.Printf("Default branch :      %s\n", *repo.DefaultBranch)

// 	protection, _, err := client.Repositories.GetBranchProtection(ctx, owner, reponame, *repo.DefaultBranch)
// 	if protection != nil {
// 		prReviews := protection.GetRequiredPullRequestReviews()
// 		if prReviews != nil {
// 			log.Printf("Branch protection (%s), requires code review :  %t\n", *repo.DefaultBranch, true)
// 			log.Printf("Branch protection (%s), approval count :        %d\n", *repo.DefaultBranch, prReviews.RequiredApprovingReviewCount)
// 		} else {
// 			log.Printf("Branch protection (%s), requires code review :  %t\n", *repo.DefaultBranch, false)
// 		}

// 		prStatusChecks := protection.GetRequiredStatusChecks()
// 		if prStatusChecks != nil {
// 			log.Printf("Branch protection (%s), branch must be up to date before merge :  %t\n", *repo.DefaultBranch, prStatusChecks.Strict)
// 			log.Printf("Branch protection (%s), status checks :  %v\n", *repo.DefaultBranch, prStatusChecks.Contexts)
// 		}

// 	} else {
// 		log.Printf("Branch protection  (%s):  None\n", *repo.DefaultBranch)
// 	}

// 	releases, err := listReleases(client, owner, reponame, maxReleases)
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("\nReleases:\n---------\n")
// 	tmpl := "%-11s %-19s %-12s %-18s %-40s\n"
// 	log.Printf(tmpl, "Status", "Published", "Tag", "Author", "Name")
// 	for i, release := range releases {
// 		if i >= maxReleases {
// 			break
// 		}
// 		status := ""
// 		if *release.Draft {
// 			status = "Draft"
// 		} else if *release.Prerelease {
// 			status = "Pre-release"
// 		} else {
// 			status = "Published"
// 		}
// 		log.Printf(tmpl, status, formatDate(release.PublishedAt), *release.TagName, *release.Author.Login, release.GetName())
// 	}

// 	tags, err := listTags(client, owner, reponame, maxTags)
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("\nTags:\n---------\n")
// 	tmpl = "%-12s %-45s\n"
// 	log.Printf(tmpl, "Name", "Commit")
// 	for i, tag := range tags {
// 		if i >= maxTags {
// 			break
// 		}
// 		log.Printf(tmpl, *tag.Name, tag.Commit.GetSHA())
// 	}

// 	return nil
// }

// func formatDate(t *github.Timestamp) string {
// 	if t == nil {
// 		return ""
// 	}
// 	return t.In(time.Local).Format("2006-01-02 15:04:05")
// }

// // returns tags in created order :-(
// func listTags(client *github.Client, owner, repo string, max int) ([]*github.RepositoryTag, error) {
// 	ctx := context.Background()
// 	opt := &github.ListOptions{PerPage: 100}
// 	// get all pages of results
// 	var allTags []*github.RepositoryTag
// 	for {
// 		releases, resp, err := client.Repositories.ListTags(ctx, owner, repo, opt)
// 		if err != nil {
// 			return allTags, err
// 		}
// 		allTags = append(allTags, releases...)
// 		if resp.NextPage == 0 {
// 			break
// 		}
// 		// Break after retrieved max. Note this function can returned a slice larger than max, because
// 		// we retrieve a page at a time
// 		if max > 0 && len(allTags) > max {
// 			break
// 		}
// 		opt.Page = resp.NextPage
// 	}
// 	return allTags, nil
// }

// func listReleases(client *github.Client, owner, repo string, max int) ([]*github.RepositoryRelease, error) {
// 	ctx := context.Background()
// 	opt := &github.ListOptions{PerPage: 100}
// 	// get all pages of results
// 	var allReleases []*github.RepositoryRelease
// 	for {
// 		releases, resp, err := client.Repositories.ListReleases(ctx, owner, repo, opt)
// 		if err != nil {
// 			return allReleases, err
// 		}
// 		allReleases = append(allReleases, releases...)
// 		if resp.NextPage == 0 {
// 			break
// 		}
// 		// Break after retrieved max. Note this function can returned a slice larger than max, because
// 		// we retrieve a page at a time
// 		if max > 0 && len(allReleases) > max {
// 			break
// 		}
// 		opt.Page = resp.NextPage
// 	}
// 	return allReleases, nil
// }

func listReposByUser(client *githubv4.Client, user string) ([]Repository, error) {
	ctx := context.Background()
	var q struct {
		User struct {
			Login        githubv4.String
			Repositories struct {
				Nodes    []Repository
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"repositories(first: 100, after: $repositoriesCursor)"` // 100 per page.
		} `graphql:"user(login: $login)"`
	}

	variables := map[string]interface{}{
		"login":              githubv4.String(user),
		"repositoriesCursor": (*githubv4.String)(nil), // Null after argument to get first page.
	}
	var allRepos []Repository
	for {
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, q.User.Repositories.Nodes...)
		if !q.User.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["repositoriesCursor"] = githubv4.NewString(q.User.Repositories.PageInfo.EndCursor)
	}
	return allRepos, nil
}

func listReposByOrg(client *githubv4.Client, org string) ([]Repository, error) {
	ctx := context.Background()
	var q struct {
		Organization struct {
			Login        githubv4.String
			Repositories struct {
				Nodes    []Repository
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"repositories(first: 100, after: $repositoriesCursor)"` // 100 per page.
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]interface{}{
		"login":              githubv4.String(org),
		"repositoriesCursor": (*githubv4.String)(nil), // Null after argument to get first page.
	}
	var allRepos []Repository
	for {
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, q.Organization.Repositories.Nodes...)
		if !q.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["repositoriesCursor"] = githubv4.NewString(q.Organization.Repositories.PageInfo.EndCursor)
	}
	return allRepos, nil
}
