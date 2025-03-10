package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/stretchr/testify/assert"
)

func TestUpdateAvailable(t *testing.T) {
	updater := NewUpdater()
	home, cleanup := TempHome()
	defer cleanup()
	updater.home = home
	output := SpyOn(updater)
	oldVersion := "v1.20.0"
	currentRelease, _, _ := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("thoughtworks/talisman"))
	updater.Check(context.Background(), oldVersion)
	assert.True(t, currentRelease.GreaterThan(oldVersion), "There is an update available for this old version.")
	assert.Equal(t, UpdateMessage("", currentRelease.Version()), output.String(), "There is an update available for this old version.")
}

func TestNoUpdateAvailableForInvalidUpdateQuery(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	invalidUpdater := Updater{client: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("/bad-repo"), home: home}
	output := SpyOn(&invalidUpdater)
	invalidUpdater.Check(context.Background(), "v1.32.0")
	assert.True(t, output.String() == "", "We should not suggest updating if there might not be an update. This simulates network errors or GitHub rate limiting.")
}

func TestNoUpdateAvailableIfNoReleaseFound(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	invalidUpdater := Updater{client: selfupdate.DefaultUpdater(), repository: selfupdate.ParseSlug("thoughtworks/thoughtworks.github.io"), home: home}
	output := SpyOn(&invalidUpdater)
	invalidUpdater.Check(context.Background(), "v0.0.0")
	assert.True(t, output.String() == "", "There is no update available if there are no releases")
}

func TestNoUpdateAvailableIfOnCurrentVersion(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	currentRelease, _, _ := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("thoughtworks/talisman"))
	updater := NewUpdater()
	updater.home = home
	output := SpyOn(updater)
	updater.Check(context.Background(), currentRelease.Version())
	assert.True(t, output.String() == "", "There is no update available if on the current version")
}

func TestNoUpdateAvailableIfUnexpectedCurrentVersion(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	updater := NewUpdater()
	updater.home = home
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

func TestUseHomeSetByScriptInstallationIfSet(t *testing.T) {
	os.Setenv("TALISMAN_HOME", "/usr/home/tester/.talisman/bin")
	defer os.Unsetenv("TALISMAN_HOME")
	scriptInstallHome := DefaultTalismanHome()
	assert.Equal(t, "/usr/home/tester/.talisman/bin", scriptInstallHome.path, "Should use directory specified by TALISMAN_HOME if set.")
}

func TestUseDirectoryInUserHome(t *testing.T) {
	os.Unsetenv("TALISMAN_HOME")
	defaultHome := DefaultTalismanHome()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, filepath.Join(homeDir, ".talisman"), defaultHome.path, "Should use $HOME/.talisman when TALISMAN_HOME is not set")
}

func TestUpdateCheckFileInTalismanHome(t *testing.T) {
	talismanHome := Home{"/usr/test/.talisman"}
	assert.Equal(t, filepath.Join("/usr/test/.talisman", "version-check"), talismanHome.updateCheckFile(), "Should look for version check file in talisman home directory")
}

func TestReturnsZeroTimeIfNoUpdateCheckFile(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	never := home.LastCheckedForUpdateAt()
	assert.Equal(t, time.Unix(0, 0), never, "Should report zero time if update check file isn't present")
}

func TestReturnsZeroTimeIfUpdateCheckFileIsMalformed(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	err := os.WriteFile(home.updateCheckFile(), []byte("not a timestamp"), 0660)
	if err != nil {
		panic(err)
	}
	never := home.LastCheckedForUpdateAt()
	assert.Equal(t, time.Unix(0, 0), never, "Should report zero time if update check file is not in the expected format")
}

func TestReturnsTimeStoredInUpdateCheckFile(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	updatedAt := time.Now().In(time.UTC).Truncate(time.Hour * 24)
	err := os.WriteFile(home.updateCheckFile(), []byte(updatedAt.Format(time.DateOnly)), 0660)
	if err != nil {
		panic(err)
	}
	lastUpdatedAt := home.LastCheckedForUpdateAt()
	assert.Equal(t, updatedAt, lastUpdatedAt, "Should read last updated time from file")
}

func TempHome() (*Home, func()) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "talisman-test-home")
	if err != nil {
		panic(err)
	}
	return &Home{tempDir}, func() { os.RemoveAll(tempDir) }
}

func TestCreatesHomeDirIfNotExists(t *testing.T) {
	home, cleanup := TempHome()
	defer cleanup()
	nonExistantHome := filepath.Join(home.path, ".talisman-missing")
	os.Setenv("TALISMAN_HOME", nonExistantHome)
	defer os.Unsetenv("TALISMAN_HOME")
	_ = DefaultTalismanHome()
	fileinfo, err := os.Stat(nonExistantHome)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			assert.Fail(t, "Should create talisman home directory if it doesn't exist")
		}
		panic(err)
	}
	assert.True(t, fileinfo.IsDir(), "Talisman home should be a directory")
}
