package main

import (
	"bytes"
	"testing"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/stretchr/testify/assert"
)

var u = GitHubClient{repoSlug: "thoughtworks/talisman"}

func TestIfUpdateAvailable(t *testing.T) {
	oldVersion, _ := semver.ParseTolerant("v1.20.0")
	updateAvailable, availableVersion := u.CanUpdateFrom(oldVersion.String())
	assert.True(t, updateAvailable, "There is an update available for this old version.")
	newVersion, _ := semver.ParseTolerant(availableVersion)
	assert.True(t, newVersion.GT(oldVersion), "The update available is a greater semantic version.")
}

func TestNoUpdateAvailableForInvalidQuery(t *testing.T) {
	invalidRepoClient := GitHubClient{repoSlug: "/bad-repo"}
	updateAvailable, _ := invalidRepoClient.CanUpdateFrom("v1.32.0")
	assert.False(t, updateAvailable, "We should not suggest updating if there might not be an update. This simulates network errors or GitHub rate limiting.")
}

func TestNoUpdateAvailableIfNoReleaseFound(t *testing.T) {
	noReleaseClient := GitHubClient{repoSlug: "thoughtworks/thoughtworks.github.io"}
	updateAvailable, _ := noReleaseClient.CanUpdateFrom("0.0.0")
	assert.False(t, updateAvailable, "There is no update available if there are no releases")
}

func TestNoUpdateAvailableIfOnCurrentVersion(t *testing.T) {
	currentRelease, _, _ := selfupdate.DetectLatest("thoughtworks/talisman")
	currentVersion := currentRelease.Version.String()
	updateAvailable, _ := u.CanUpdateFrom(currentVersion)
	assert.False(t, updateAvailable, "There is no update available if on the current version")
}

func TestNoUpdateIfUnexpectedCurrentVersion(t *testing.T) {
	updateAvailable, _ := u.CanUpdateFrom("Local dev version")
	assert.False(t, updateAvailable, "There is no update available if not on a published version")
}

func TestPrintsMessageWhenUpdateAvailable(t *testing.T) {
	newVersion := "v3.25.1"
	var output bytes.Buffer
	updater := Updater{client: &EagerClient{nextVersion: newVersion}, output: &output}
	updater.Check("")
	assert.Equal(t, UpdateMessage(newVersion), output.String())
}

type EagerClient struct {
	nextVersion string
}

func (u *EagerClient) CanUpdateFrom(_ string) (bool, string) {
	return true, u.nextVersion
}

func (u *EagerClient) Update(_ string) error {
	return nil
}
