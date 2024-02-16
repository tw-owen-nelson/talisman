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

func NewUpdater() *UpdateManager {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums"}})
	if err != nil {
		panic(err)
	}
	repository := selfupdate.ParseSlug("thoughtworks/talisman")
	return &UpdateManager{
		updater:    updater,
		repository: &repository,
		output:     os.Stdout,
	}
}

type UpdateManager struct {
	updater    *selfupdate.Updater
	repository selfupdate.Repository
	output     io.Writer
}

func (um *UpdateManager) CanUpdateFrom(ctx context.Context, currentVersion string) {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return
	}
	release, _, err := um.updater.DetectLatest(ctx, um.repository)
	if err != nil || release == nil {
		return
	}
	if release.GreaterThan(currentVersion) {
		fmt.Fprint(um.output, UpdateMessage(release.Version()))
	}
}

func (um *UpdateManager) Update(ctx context.Context, currentVersion string) error {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return fmt.Errorf("unexpected value for currently installed version: %s", currentVersion)
	}
	_, err := um.updater.UpdateSelf(ctx, currentVersion, um.repository)
	return err
}
