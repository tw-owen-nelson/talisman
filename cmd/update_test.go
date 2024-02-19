package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/stretchr/testify/assert"
)

func TestUpdateAvailable(t *testing.T) {
	updater := NewUpdater()
	output := SpyOn(updater)
	oldVersion := "v1.20.0"
	currentRelease, _, _ := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("thoughtworks/talisman"))
	updater.Check(context.Background(), oldVersion)
	assert.True(t, currentRelease.GreaterThan(oldVersion), "There is an update available for this old version.")
	assert.Equal(t, UpdateMessage("", currentRelease.Version()), output.String(), "There is an update available for this old version.")
}

func TestNoUpdateAvailableForInvalidUpdateQuery(t *testing.T) {
	invalidUpdater := Updater{client: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("/bad-repo")}
	output := SpyOn(&invalidUpdater)
	invalidUpdater.Check(context.Background(), "v1.32.0")
	assert.True(t, output.String() == "", "We should not suggest updating if there might not be an update. This simulates network errors or GitHub rate limiting.")
}

func TestNoUpdateAvailableIfNoReleaseFound(t *testing.T) {
	invalidUpdater := Updater{client: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("thoughtworks/thoughtworks.github.io")}
	output := SpyOn(&invalidUpdater)
	invalidUpdater.Check(context.Background(), "v0.0.0")
	assert.True(t, output.String() == "", "There is no update available if there are no releases")
}

func TestNoUpdateAvailableIfOnCurrentVersion(t *testing.T) {
	currentRelease, _, _ := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("thoughtworks/talisman"))
	updater := NewUpdater()
	output := SpyOn(updater)
	updater.Check(context.Background(), currentRelease.Version())
	assert.True(t, output.String() == "", "There is no update available if on the current version")
}

func TestNoUpdateAvailableIfUnexpectedCurrentVersion(t *testing.T) {
	updater := NewUpdater()
	output := SpyOn(updater)
	updater.Check(context.Background(), "Local dev version")
	assert.True(t, output.String() == "", "There is no update available if not on a published version")
}

func SpyOn(updater *Updater) *bytes.Buffer {
	var output bytes.Buffer
	updater.output = &output
	return &output
}

func TestSuggestTalismanUpgradeForNativeInstall(t *testing.T) {
	nativeInstallation, cleanUp := InstallTalisman()
	defer cleanUp()
	talismanUpgradeMessage := `Talisman version v1.32.0 is available.
To upgrade, run:

	talisman -u

`
	assert.Equal(t, talismanUpgradeMessage, UpdateMessage(nativeInstallation, "v1.32.0"), "Should give homebrew command if installed by homebrew")
}

func TestSuggestTalismanUpgradeIfUnknownPath(t *testing.T) {
	talismanUpgradeMessage := `Talisman version v1.32.0 is available.
To upgrade, run:

	talisman -u

`
	assert.Equal(t, talismanUpgradeMessage, UpdateMessage("", "v1.32.0"), "Should give homebrew command if installed by homebrew")
}

func TestSuggestBrewUpgradeForBrewInstall(t *testing.T) {
	brewTalisman, cleanUp := BrewInstallTalisman()
	defer cleanUp()

	brewUpgradeMessage := `Talisman version v1.32.0 is available.
To upgrade, run:

	brew upgrade talisman

`
	assert.Equal(t, brewUpgradeMessage, UpdateMessage(brewTalisman, "v1.32.0"), "Should give homebrew command if installed by homebrew")
}

func InstallTalisman() (string, func()) {
	usrLocalBin, _ := os.MkdirTemp(os.TempDir(), "talisman-test-bin")
	nativeInstallation := filepath.Join(usrLocalBin, "talisman")
	os.WriteFile(nativeInstallation, []byte(""), 0755)
	return nativeInstallation, func() { os.RemoveAll(usrLocalBin) }
}

func BrewInstallTalisman() (string, func()) {
	brewHome, _ := os.MkdirTemp(os.TempDir(), "talisman-test-homebrew")

	brewCellar := filepath.Join(brewHome, "Cellar")
	_ = os.Mkdir(brewCellar, 0755)

	brewInstallation := filepath.Join(brewCellar, "talisman")
	os.WriteFile(brewInstallation, []byte(""), 0755)

	brewBins := filepath.Join(brewHome, "bin")
	_ = os.Mkdir(brewBins, 0755)

	brewTalisman := filepath.Join(brewBins, "talisman")
	_ = os.Symlink(brewInstallation, brewTalisman)

	return brewTalisman, func() { os.RemoveAll(brewHome) }
}
