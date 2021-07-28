package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

func main() {
	log.SetFlags(0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	args := runArgs{
		Template:  os.Getenv("INPUT_TEMPLATE-FILE"),
		Variables: os.Getenv("INPUT_VARIABLES-FILE"),
	}
	flag.StringVar(&args.Template, "t", args.Template, "path to the template, see: https://pkg.go.dev/text/template")
	flag.StringVar(&args.Variables, "v", args.Variables, "path to JSON mapping of variables to use in template, see:\nhttps://pkg.go.dev/encoding/json#Unmarshal")
	flag.Parse()
	if err := run(ctx, args); err != nil {
		log.Fatal(err)
	}
}

type runArgs struct {
	Template  string
	Variables string
}

func run(ctx context.Context, args runArgs) error {
	if args.Template == "" || args.Variables == "" {
		return errors.New("need both template and variables")
	}
	eventFile := os.Getenv("GITHUB_EVENT_PATH")
	if eventFile == "" {
		return errors.New("GITHUB_EVENT_PATH is empty")
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("INPUT_GITHUB-TOKEN")
	}
	if githubToken == "" {
		return errors.New("GITHUB_TOKEN must be set")
	}
	if eventFile == "" {
		return errors.New("no event file profided")
	}

	tpl, err := template.ParseFiles(args.Template)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(args.Variables)
	if err != nil {
		return err
	}
	var vars map[string]interface{}
	if err := json.Unmarshal(data, &vars); err != nil {
		return err
	}
	body := new(bytes.Buffer)
	if err := tpl.Execute(body, vars); err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	data, err = os.ReadFile(eventFile)
	if err != nil {
		return err
	}
	event := struct {
		PR *github.PullRequest `json:"pull_request"`
	}{}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	if event.PR == nil {
		return errors.New("payload is not a pull request event")
	}
	pr := event.PR
	if pr.GetState() != "open" {
		return errors.New("event is not an open PR")
	}
	if pr.Number == nil {
		return errors.New("nil pr.Number")
	}

	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})))
	var owner string
	{
		ow := pr.GetBase().GetRepo().GetOwner()
		owner = ow.GetLogin()
	}
	repo := pr.GetBase().GetRepo().GetName()

	text := body.String()
	comment := &github.IssueComment{
		Body: &text,
	}
	comment, _, err = client.Issues.CreateComment(ctx, owner, repo, *pr.Number, comment)
	if err != nil {
		return err
	}
	fmt.Printf("::set-output name=comment-id::%d\n", comment.GetID())
	fmt.Printf("::set-output name=comment-url::%s\n", comment.GetHTMLURL())
	return nil
}
