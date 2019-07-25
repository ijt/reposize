// Command reposize takes a newline-separated list of github repos on stdin,
// clones them, computes their size, deletes them, and outputs the total
// size in bytes on stdout.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

func main() {
	sizeBytes := 0
	n := 0
	s := bufio.NewScanner(os.Stdin)
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
}

func repoSize(repo string) (int, error) {
	// Clone the repo.
	out, err := exec.Command("git", "clone", repo).CombinedOutput()
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
