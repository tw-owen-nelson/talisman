package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/blang/semver"
	"github.com/creativeprojects/go-selfupdate"
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

func (u *GitHubClient) CanUpdateFrom(currentVersion string) (bool, string) {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return false, ""
	}
	release, _, err := selfupdate.DetectLatest(context.TODO(), selfupdate.ParseSlug(u.repoSlug))
	if err != nil || release == nil {
		return false, ""
	}
	return release.GreaterThan(currentVersion), release.Version()
}

func (u *GitHubClient) Update(currentVersion string) error {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return fmt.Errorf("unexpected value for currently installed version: %s", currentVersion)
	}
	_, err := selfupdate.UpdateSelf(context.TODO(), currentVersion, selfupdate.ParseSlug(u.repoSlug))
	return err
}
