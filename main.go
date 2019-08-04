package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Luzifer/rconfig/v2"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	cfg = struct {
		DryRun           bool   `flag:"dry-run,n" default:"false" description:"Only report actions to be done, don't execute them"`
		GiteaToken       string `flag:"gitea-token" default:"" description:"Token to interact with Gitea instance" validate:"nonzero"`
		GiteaURL         string `flag:"gitea-url" default:"" description:"URL of the Gitea instance" validate:"nonzero"`
		GithubToken      string `flag:"github-token" default:"" description:"Github access token" validate:"nonzero"`
		LogLevel         string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		MappingFile      string `flag:"mapping-file" default:"" description:"File containing several mappings to execute in one run"`
		MigrateArchived  bool   `flag:"migrate-archived" default:"false" description:"Create migrations for archived repos"`
		MigrateForks     bool   `flag:"migrate-forks" default:"false" description:"Create migrations for forked repos"`
		MigratePrivate   bool   `flag:"migrate-private" default:"true" description:"Migrate private repos (the given Github Token will be entered as sync credential!)"`
		NoMirror         bool   `flag:"no-mirror" default:"false" description:"Do not enable mirroring but instad do a one-time clone"`
		SourceExpression string `flag:"source-expression" default:"" description:"Regular expression to match the full name of the source repo (i.e. '^Luzifer/.*$')"`
		TargetUser       int64  `flag:"target-user" default:"0" description:"ID of the User / Organization in Gitea to assign the repo to"`
		TargetUserName   string `flag:"target-user-name" default:"" description:"Username of the given ID (to check whether repo already exists)"`
		VersionAndExit   bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	mappings *mapFile

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("create-gitea-migration %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	mappings = newMapFile()

	switch {

	case cfg.MappingFile != "":
		m, err := loadMapFile(cfg.MappingFile)
		if err != nil {
			log.WithError(err).Fatal("Unable to load mappings file")
		}
		mappings = m

	case cfg.SourceExpression != "" && cfg.TargetUserName != "" && cfg.TargetUser > 0:
		mappings.Mappings = []mapping{{
			SourceExpression: cfg.SourceExpression,
			TargetUser:       cfg.TargetUser,
			TargetUserName:   cfg.TargetUserName,
		}}

	default:
		log.Fatal("No mappings are defined")

	}

	log.WithFields(log.Fields{
		"dry-run":         cfg.DryRun,
		"map_definitions": len(mappings.Mappings),
		"version":         version,
	}).Info("create-gitea-migration started")

	log.Info("Collecting source repos...")
	repos, err := fetchGithubRepos()
	if err != nil {
		log.WithError(err).Fatal("Failed to fetch repos")
	}

	log.Info("Creating target repos...")
	for _, r := range repos {
		if err := giteaCreateMigration(r); err != nil {
			log.WithError(err).Error("Unable to create mirror")
		}
	}
}

func fetchGithubRepos() ([]*github.Repository, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// get all pages of results
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to list repos")
		}

		for _, r := range repos {
			if !mappings.MappingAvailable(*r.FullName) {
				log.WithField("repo", *r.FullName).Debug("Skip: No mapping matches repo name")
				continue
			}

			if !cfg.MigrateArchived && boolFromPtr(r.Archived) {
				log.WithField("repo", *r.FullName).Debug("Skip: Archived")
				continue
			}

			if !cfg.MigrateForks && boolFromPtr(r.Fork) {
				log.WithField("repo", *r.FullName).Debug("Skip: Fork")
				continue
			}

			allRepos = append(allRepos, r)
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func giteaCreateMigration(gr *github.Repository) error {
	repoMapping := mappings.GetMapping(*gr.FullName)

	logger := log.WithFields(log.Fields{
		"private":     boolFromPtr(gr.Private),
		"source_repo": strFromPtr(gr.FullName),
		"target_repo": strings.Join([]string{repoMapping.TargetUserName, strFromPtr(gr.Name)}, "/"),
	})

	req, _ := http.NewRequest(http.MethodGet, giteaURL(strings.Join([]string{"api/v1/repos", repoMapping.TargetUserName, strFromPtr(gr.Name)}, "/")), nil)
	req.Header.Set("Authorization", "token "+cfg.GiteaToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Unable to create repo in Gitea")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logger.Info("Repo already exists, no action required")
		return nil
	}

	cmr := createMigrationRequestFromGithubRepo(gr, repoMapping)

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(cmr); err != nil {
		return errors.Wrap(err, "Unable to marshal creation request")
	}

	if cfg.DryRun {
		logger.Warn("Repo not found, will be created in real run (dry-run enabled)")
		return nil
	}

	logger.Info("Repo not found, creating")

	req, _ = http.NewRequest(http.MethodPost, giteaURL("api/v1/repos/migrate"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+cfg.GiteaToken)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Unable to create repo in Gitea")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("Unable to create repo in Gitea: Status %d: %s", resp.StatusCode, body)
	}

	return nil
}

func giteaURL(path string) string {
	return strings.Join([]string{
		strings.TrimRight(cfg.GiteaURL, "/"),
		strings.TrimLeft(path, "/"),
	}, "/")
}
