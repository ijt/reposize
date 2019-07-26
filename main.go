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
var keepFlag = flag.Bool("k", false, "whether to keep repo dirs rather than deleting them")

func main() {
	if err := reposize(); err != nil {
		log.Fatal(err)
	}
}

func reposize() error {
	flag.Parse()
	n := 0
	s := bufio.NewScanner(os.Stdin)
	td, err := ioutil.TempDir("", "reposize")
	if err := os.Chdir(td); err != nil {
		return errors.Wrapf(err, "cd'ing into temp dir %s", td)
	}
	if err != nil {
		return errors.Wrap(err, "making temp dir")
	}
	if *verboseFlag {
		log.Printf("working in %s", td)
	}
	for s.Scan() {
		n++
		r := s.Text()
		if *verboseFlag {
			log.Printf("sizing repo %d: %s", n, r)
		}
		sb, err := sizeOfOneRepo(r)
		if err != nil {
			fmt.Printf("%d,%s\n", sb, r)
			continue
		}
	}
	return nil
}

var ghrx = regexp.MustCompile(`^github\.com/`)

func sizeOfOneRepo(repo string) (int, error) {
	// Clone the repo.
	out, err := exec.Command("git", "clone", fmt.Sprintf("https://%s.git", repo)).CombinedOutput()
	if err != nil {
		return 0, errors.Wrapf(err, "clone failed: %s", out)
	}

	// Add the size of the repo.
	d := filepath.Base(repo)
	sb, err := dirSizeBytes(d)
	if err != nil {
		return 0, errors.Wrap(err, "computing size of repo")
	}

	if !*keepFlag {
		// Delete the repo.
		if err := os.RemoveAll(d); err != nil {
			return 0, errors.Wrapf(err, "removing %s", d)
		}
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
