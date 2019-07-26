// Command reposize takes a newline-separated list of github repos on stdin,
// clones them, computes their size, deletes them, and outputs to stdout
// a CSV table with rows of sizebytes,repo.

package main // import "github.com/ijt/reposize"

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

var verboseFlag = flag.Bool("v", false, "whether to log verbosely")

func main() {
	if err := reposize(); err != nil {
		log.Fatal(err)
	}
}

func reposize() error {
	flag.Parse()
	n := 0
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		n++
		r := s.Text()
		if *verboseFlag {
			log.Printf("sizing repo %d: %s", n, r)
		}
		sb, err := sizeOfOneRepo(r)
		if err != nil {
			log.Printf("error sizing repo %s: %v", r, err)
			continue
		}
		fmt.Printf("%d,%s\n", sb, r)
	}
	return nil
}

var ghrx = regexp.MustCompile(`^github\.com/`)

func sizeOfOneRepo(repo string) (int, error) {
	td, err := ioutil.TempDir("", "reposize")
	if err != nil {
		return 0, errors.Wrap(err, "making temp dir")
	}

	// Clone the repo.
	d := filepath.Join(td, filepath.Base(repo))
	cmd := exec.Command("git", "clone", fmt.Sprintf("https://%s.git", repo), d)
	cmd.Env = []string{"GIT_TERMINAL_PROMPT=0"}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, errors.Wrapf(err, "clone failed: %s", out)
	}

	// Add the size of the repo.
	sb, err := dirSizeBytes(d)
	if err != nil {
		return 0, errors.Wrap(err, "computing size of repo")
	}

	// Delete the repo.
	if err := os.RemoveAll(td); err != nil {
		return 0, errors.Wrapf(err, "removing %s", td)
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
