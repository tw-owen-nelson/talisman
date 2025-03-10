package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/creativeprojects/go-selfupdate"
	log "github.com/sirupsen/logrus"
)

const (
	MessageTemplate = `Talisman version %s is available.
To upgrade, run:

	%s

`
	SCRIPT_INSTALL_HOME_VAR = "TALISMAN_HOME"
	UPDATE_CHECK_FILE       = "version-check"
	TIME_FORMAT             = time.DateOnly
)

type Updater struct {
	client     *selfupdate.Updater
	repository selfupdate.Repository
	output     io.Writer
	home       *Home
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
		home:       DefaultTalismanHome(),
	}
}

func (u *Updater) Check(ctx context.Context, currentVersion string) {
	if _, err := semver.ParseTolerant(currentVersion); err != nil {
		return
	}
	if time.Since(u.home.LastCheckedForUpdateAt()).Hours() < 7*24 {
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
	} else {
		u.home.RecordCheckedForUpdateAt(time.Now())
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

type Home struct {
	path string
}

func DefaultTalismanHome() *Home {
	var home *Home
	if legacy_home := os.Getenv(SCRIPT_INSTALL_HOME_VAR); legacy_home != "" {
		home = &Home{legacy_home}
	} else {
		user_home, err := os.UserHomeDir()
		if err != nil {
			panic("no user home??")
		}
		home = &Home{filepath.Join(user_home, ".talisman")}
	}
	home.Init()
	return home
}

func (h *Home) Init() {
	info, err := os.Stat(h.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err := os.MkdirAll(h.path, 0755)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	if !info.IsDir() {
		log.Fatalf("Talisman home directory '%s' already exists and is not a directory", h.path)
	}
}

func (h *Home) LastCheckedForUpdateAt() time.Time {
	timestamp, err := os.ReadFile(h.updateCheckFile())
	if err != nil {
		return time.Unix(0, 0)
	}
	t, err := time.Parse(time.DateOnly, string(timestamp))
	if err != nil {
		return time.Unix(0, 0)
	}
	return t
}

func (h *Home) RecordCheckedForUpdateAt(now time.Time) {
	checkedAt := now.Format(TIME_FORMAT)
	os.WriteFile(h.updateCheckFile(), []byte(checkedAt), 0660)
}

func (h *Home) updateCheckFile() string {
	return filepath.Join(h.path, UPDATE_CHECK_FILE)
}
