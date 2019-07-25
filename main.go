// Command reposize takes a newline-separated list of github repos on stdin,
// clones them, computes their size, deletes them, and outputs the total
// size in bytes on stdout.

package main  // import "github.com/ijt/reposize"

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

func main() {
	if err := reposize(); err != nil {
		log.Fatal(err)
	}
}

func reposize() error {
	sizeBytes := 0
	n := 0
	s := bufio.NewScanner(os.Stdin)
	td, err := ioutil.TempDir("", "reposize")
	if err := os.Chdir(td); err != nil {
		return errors.Wrapf(err, "cd'ing into temp dir %s", td)
	}
	if err != nil {
		return errors.Wrap(err, "making temp dir")
	}
	for s.Scan() {
		r := s.Text()
		log.Printf("sizing %s", r)
		sb, err := repoSize(r)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		sizeBytes += sb
		n++
	}
	fmt.Printf("%d bytes (%.3fG) in %d repos\n", sizeBytes, float64(sizeBytes)/(1024.0*1024.0*1024.0), n)
	return nil
}

var ghrx = regexp.MustCompile(`^github\.com/`)

func repoSize(repo string) (int, error) {
	// Clone the repo.
	repo2 := ghrx.ReplaceAllString(repo, "git@github.com:")
	out, err := exec.Command("git", "clone", repo2).CombinedOutput()
	if err != nil {
		return 0, errors.Wrapf(err, "clone failed: %s", out)
	}

	// Add the size of the repo.
	d := filepath.Base(repo)
	sb, err := dirSizeBytes(d)
	if err != nil {
		return 0, errors.Wrap(err, "computing size of repo")
	}

	// Delete the repo.
	if err := os.RemoveAll(d); err != nil {
		return 0, errors.Wrapf(err, "removing %s", d)
	}

	return sb, nil
}

func dirSizeBytes(d string) (int, error) {
	sb := 0
	err := filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		sb += int(info.Size())
		return nil
	})
	if err != nil {
		return 0, errors.Wrapf(err, "walking from %s", d)
	}
	return sb, nil
}
