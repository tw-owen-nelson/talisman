package main

import (
	"fmt"
	"io"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func NewUpdater() *Updater {
	return &Updater{
		client: &GitHubClient{repoSlug: "thoughtworks/talisman"},
		output: os.Stdout,
	}
}

func UpdateMessage(newVersion string) string {
	return fmt.Sprintf("Talisman version %s is available.\n", newVersion)
}

type Updater struct {
	client UpdateClient
	output io.Writer
}

type UpdateClient interface {
	CanUpdateFrom(string) (bool, string)
}

type GitHubClient struct {
	repoSlug string
}

func (u *GitHubClient) CanUpdateFrom(current string) (bool, string) {
	currentVersion, err := semver.ParseTolerant(current)
	if err != nil {
		return false, ""
	}
	release, _, err := selfupdate.DetectLatest(u.repoSlug)
	if err != nil || release == nil {
		return false, ""
	}
	return release.Version.GT(currentVersion), release.Version.String()
}

func (u *Updater) Check(currentVersion string) {
	updateAvailable, newVersion := u.client.CanUpdateFrom(currentVersion)
	if updateAvailable {
		fmt.Fprint(u.output, UpdateMessage(newVersion))
	}
}
