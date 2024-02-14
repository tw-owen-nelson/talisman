package main

import (
	"fmt"
	"io"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func UpdateMessage(newVersion string) string {
	return fmt.Sprintf(
		`Talisman version %s is available.
To upgrade, run:

	talisman -u

`, newVersion)
}

func NewUpdater() *Updater {
	return &Updater{
		client: &GitHubClient{repoSlug: "thoughtworks/talisman"},
		output: os.Stdout,
	}
}

type Updater struct {
	client UpdateClient
	output io.Writer
}

type UpdateClient interface {
	CanUpdateFrom(string) (bool, string)
	Update(string) error
}

func (u *Updater) Check(current string) {
	updateAvailable, newVersion := u.client.CanUpdateFrom(current)
	if updateAvailable {
		fmt.Fprint(u.output, UpdateMessage(newVersion))
	}
}

func (u *Updater) Update(current string) error {
	return u.client.Update(current)
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

func (u *GitHubClient) Update(current string) error {
	currentVersion, err := semver.ParseTolerant(current)
	if err != nil {
		return fmt.Errorf("unexpected value for currently installed version: %s", current)
	}
	_, err = selfupdate.UpdateSelf(currentVersion, u.repoSlug)
	return err
}
