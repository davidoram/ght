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
)

const (
	helptext = `
ght is a tool for interacting with github from a script

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
	orgPtr := reposCommand.String("o", "", "Specify the organisation")

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
				err = doListRepos(reposCommand, orgPtr, true)

			default:
				log.Printf("Unknown command '%s'", os.Args[2])
				log.Println(helptext)
				os.Exit(1)

			}
		} else {
			log.Println(helptext)
		}
	case "repos":
		reposCommand.Parse(os.Args[2:])
		err = doListRepos(reposCommand, orgPtr, false)
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

func doListRepos(flags *flag.FlagSet, org *string, displayHelp bool) error {

	helptext := `
ght repos 		List github repositories for an organisation

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

	if *org == "" {
		log.Println(helptext)
		flags.PrintDefaults()
		return fmt.Errorf("Invalid arguments. Missing organisation argument")
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	allRepos, err := listRepos(client, *org)
	if err != nil {
		return err
	}
	for _, r := range allRepos {
		log.Printf("%s\n", *r.FullName)
	}
	return nil
}

func listRepos(client *github.Client, org string) ([]*github.Repository, error) {
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
