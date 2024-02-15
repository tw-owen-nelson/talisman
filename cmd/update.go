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
	CanUpdateFrom(context.Context, string) (bool, string)
	Update(context.Context, string) error
}

func (u *Updater) Check(ctx context.Context, current string) {
	updateAvailable, newVersion := u.client.CanUpdateFrom(ctx, current)
	if updateAvailable {
		fmt.Fprint(u.output, UpdateMessage(newVersion))
	}
}

func (u *Updater) Update(ctx context.Context, current string) error {
	return u.client.Update(ctx, current)
}

type GitHubClient struct {
	repoSlug string
}

func (u *GitHubClient) CanUpdateFrom(ctx context.Context, currentVersion string) (bool, string) {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return false, ""
	}
	release, _, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(u.repoSlug))
	if err != nil || release == nil {
		return false, ""
	}
	return release.GreaterThan(currentVersion), release.Version()
}

func (u *GitHubClient) Update(ctx context.Context, currentVersion string) error {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return fmt.Errorf("unexpected value for currently installed version: %s", currentVersion)
	}
	_, err := selfupdate.UpdateSelf(ctx, currentVersion, selfupdate.ParseSlug(u.repoSlug))
	return err
}
