package main

/* ght is the 'GitHub Tool', s read only tool for displaying information about github repos
 *
 * See:
 * - https://developer.github.com/v4/explorer/
 * - https://github.com/shurcooL/githubv4
 */

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

// Repository is the GitHub representation of a repository
type Repository struct {
	Name          string
	NameWithOwner string
}

// Release is the GitHub representation of a release, see https://help.github.com/categories/releases/
type Release struct {
	Author struct {
		Login githubv4.String
	}
	PublishedAt  githubv4.DateTime
	Name         githubv4.String
	Description  githubv4.String
	IsDraft      githubv4.Boolean
	IsPrerelease githubv4.Boolean
	Url          githubv4.URI
	Tag          struct {
		Name githubv4.String
	}
}

// QueryReposByUser is a query that returns all repositories accessable by a specific user login
type QueryReposByUser struct {
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

// QueryReposByUser is a query that returns all repositories owned by a specific organisation
type QueryReposByOrg struct {
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

// QueryRepoDetail is a query that returns detailed information for a single repository
type QueryRepoDetail struct {
	Repository struct {
		NameWithOwner    githubv4.String
		DefaultBranchRef struct {
			Name githubv4.String
		}
		BranchProtectionRules struct {
			Nodes []struct {
				MatchingRefs struct {
					Nodes []struct {
						Name githubv4.String
					}
				} `graphql:"matchingRefs(first: 10)"`
				RequiresApprovingReviews     githubv4.Boolean
				RequiredApprovingReviewCount githubv4.Int
				RequiresStatusChecks         githubv4.Boolean
				RequiredStatusCheckContexts  []githubv4.String
			}
		} `graphql:"branchProtectionRules(first: 10)"`
		Releases struct {
			Nodes []Release
		} `graphql:"releases(first: $maxReleases, orderBy: {field: CREATED_AT, direction: DESC})"`
		Tags struct {
			Edges []struct {
				Tag struct {
					Name   githubv4.String
					Target struct {
						Oid githubv4.String
					}
				} `graphql:"tag: node()"`
			}
		} `graphql:"tags: refs(refPrefix: $tagPrefix, last: $maxTags, orderBy: {field: TAG_COMMIT_DATE, direction: ASC})"`
	} `graphql:"repository(owner: $owner, name: $name)"`
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
	repoCommand := flag.NewFlagSet("repo", flag.ExitOnError)
	maxReleasesPtr := reposCommand.Int("maxr", 20, "Specify the maximum number of Releases to display")
	maxTagsPtr := reposCommand.Int("maxt", 20, "Specify the maximum number of Tags to display")

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
				err = doRepo(repoCommand, *maxReleasesPtr, *maxTagsPtr, true)

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
		err = doRepo(repoCommand, *maxReleasesPtr, *maxTagsPtr, false)

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

/* doListRepos displays information about all repos for a user, or for an organisation */
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

/* doRepo displays information about one repo */
func doRepo(flags *flag.FlagSet, maxReleases, maxTags int, displayHelp bool) error {

	helptext := `
ght repo 		Summarise a given repository

Usage:

	mdd repo owner/repo

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

	ownerRepo := strings.Split(os.Args[2], "/")
	if len(ownerRepo) != 2 {
		return fmt.Errorf("Error parsing %s as 'owner/repo'", os.Args[2])
	}
	owner := ownerRepo[0]
	reponame := ownerRepo[1]

	ctx := context.Background()
	var q QueryRepoDetail
	variables := map[string]interface{}{
		"owner":       githubv4.String(owner),
		"name":        githubv4.String(reponame),
		"maxReleases": githubv4.Int(maxReleases),
		"maxTags":     githubv4.Int(maxTags),
		"tagPrefix":   githubv4.String("refs/tags/"),
	}

	err = client.Query(ctx, &q, variables)
	if err != nil {
		return err
	}

	log.Printf("Full name :           %s\n", q.Repository.NameWithOwner)
	log.Printf("Default branch :      %s\n", q.Repository.DefaultBranchRef.Name)
	if len(q.Repository.BranchProtectionRules.Nodes) == 0 {
		log.Printf("Branch protection :   None\n")
	} else {
		for _, bpr := range q.Repository.BranchProtectionRules.Nodes {
			for _, b := range bpr.MatchingRefs.Nodes {
				log.Printf("\nBranch protection ('%s') : Enabled\n", b.Name)
				log.Printf("  - approving review       :  %t\n", bpr.RequiresApprovingReviews)
				log.Printf("  - approving review count :  %d\n", bpr.RequiredApprovingReviewCount)
				log.Printf("  - status check           :  %t\n", bpr.RequiresStatusChecks)
				log.Printf("  - status check contexts  :  %v\n", bpr.RequiredStatusCheckContexts)
			}
		}
	}

	log.Printf("\nReleases:\n---------\n")
	tmpl := "%-11s %-19s %-12s %-18s %-40s\n"
	log.Printf(tmpl, "Status", "Published", "Tag", "Author", "Name")
	for i, r := range q.Repository.Releases.Nodes {
		if i >= maxReleases {
			break
		}
		log.Printf(tmpl, formatStatus(r), formatDate(r.PublishedAt), r.Tag.Name, r.Author.Login, r.Name)
	}

	log.Printf("\nTags:\n-----\n")
	tmpl = "%-70s %s\n"
	log.Printf(tmpl, "Tag", "Sha")
	for i, t := range q.Repository.Tags.Edges {
		if i >= maxTags {
			break
		}
		log.Printf(tmpl, t.Tag.Name, t.Tag.Target.Oid)
	}
	return nil
}

func formatStatus(r Release) string {
	if r.IsDraft {
		return "Draft"
	} else if r.IsPrerelease {
		return "Pre-release"
	}
	return "Published"
}

func formatDate(t githubv4.DateTime) string {
	return t.In(time.Local).Format("2006-01-02 15:04:05")
}

func listReposByUser(client *githubv4.Client, user string) ([]Repository, error) {
	ctx := context.Background()
	var q QueryReposByUser
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
	var q QueryReposByOrg
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
