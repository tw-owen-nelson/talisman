package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/creativeprojects/go-selfupdate"
	log "github.com/sirupsen/logrus"
)

const MessageTemplate = `Talisman version %s is available.
To upgrade, run:

	%s

`

type Updater struct {
	client     *selfupdate.Updater
	repository selfupdate.Repository
	output     io.Writer
}

func NewUpdater() *Updater {
	client, err := selfupdate.NewUpdater(selfupdate.Config{Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums"}})
	if err != nil {
		panic(err)
	}
	repository := selfupdate.ParseSlug("thoughtworks/talisman")
	return &Updater{
		client:     client,
		repository: &repository,
		output:     os.Stdout,
	}
}

func (u *Updater) Check(ctx context.Context, currentVersion string) {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return
	}
	release, _, err := u.client.DetectLatest(ctx, u.repository)
	if err != nil || release == nil {
		return
	}
	if release.GreaterThan(currentVersion) {
		executable, err := os.Executable()
		if err != nil {
			executable = ""
		}
		fmt.Fprint(u.output, UpdateMessage(executable, release.Version()))
	}
}

func (u *Updater) Update(ctx context.Context, currentVersion string) int {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		log.Errorf("unexpected value for currently installed version: %s", currentVersion)
		return EXIT_FAILURE
	}
	executable, _ := os.Executable()
	if IsHomebrewInstall(executable) {
		log.Error("Detected homebrew-managed talisman install. Please upgrade through homebrew.")
		return EXIT_FAILURE
	}
	updated, err := u.client.UpdateSelf(ctx, currentVersion, u.repository)
	if err != nil {
		log.Error(err)
		return EXIT_FAILURE
	}
	fmt.Fprintf(u.output, "Talisman updated to %s\n", updated.Version())
	return EXIT_SUCCESS
}

func UpdateMessage(path string, newVersion string) string {
	upgradeCommand := "talisman -u"
	if IsHomebrewInstall(path) {
		upgradeCommand = "brew upgrade talisman"
	}
	return fmt.Sprintf(MessageTemplate, newVersion, upgradeCommand)
}

func IsHomebrewInstall(path string) bool {
	link, _ := os.Readlink(path)
	return link != "" && strings.Contains(link, "Cellar")
}
