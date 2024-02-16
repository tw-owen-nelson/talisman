package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/stretchr/testify/assert"
)

func TestIfUpdateAvailable(t *testing.T) {
	var output bytes.Buffer
	um := UpdateManager{updater: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("thoughtworks/talisman"), output: &output}
	oldVersion := "v1.20.0"
	currentRelease, _, _ := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("thoughtworks/talisman"))
	um.Check(context.Background(), oldVersion)
	assert.True(t, currentRelease.GreaterThan(oldVersion), "There is an update available for this old version.")
	assert.Equal(t, UpdateMessage(currentRelease.Version()), output.String(), "There is an update available for this old version.")
}
func TestNoUpdateAvailableForInvalidRepository(t *testing.T) {
	var output bytes.Buffer
	invalidUm := UpdateManager{updater: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("/bad-repo"), output: &output}
	invalidUm.Check(context.Background(), "v1.32.0")
	assert.True(t, output.String() == "", "We should not suggest updating if there might not be an update. This simulates network errors or GitHub rate limiting.")
}

func TestNoUpdateAvailableIfNoReleaseFound(t *testing.T) {
	var output bytes.Buffer
	invalidUm := UpdateManager{updater: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("thoughtworks/thoughtworks.github.io"), output: &output}
	invalidUm.Check(context.Background(), "v0.0.0")
	assert.True(t, output.String() == "", "There is no update available if there are no releases")
}

func TestNoUpdateAvailableIfOnCurrentVersion(t *testing.T) {
	currentRelease, _, _ := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("thoughtworks/talisman"))
	var output bytes.Buffer
	um := UpdateManager{updater: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("thoughtworks/talisman"), output: &output}
	um.Check(context.Background(), currentRelease.Version())
	assert.True(t, output.String() == "", "There is no update available if on the current version")
}

func TestNoUpdateIfUnexpectedCurrentVersion(t *testing.T) {
	var output bytes.Buffer
	um := UpdateManager{updater: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("thoughtworks/talisman"), output: &output}
	um.Check(context.Background(), "Local dev version")
	assert.True(t, output.String() == "", "There is no update available if not on a published version")

}
