package pidfile

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type PFile struct {
	filename string
	pid      int
}

func New(filename string) *PFile {
	return &PFile{
		filename: filename,
	}
}

func (pf *PFile) Create() error {
	if err := pf.Validate(); err != nil {
		return err
	}

	if pf.pid != 0 {
		if pf.pid == os.Getpid() {
			return nil
		}
		return errors.New(fmt.Sprintf("Already running on PID %d (or pid file '%s' is stale)", pf.pid, pf.filename))
	}

	pf.pid = os.Getpid()

	path := filepath.Dir(pf.filename)
	f, err := ioutil.TempFile(path, "tmp")
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(fmt.Sprintf("%d", pf.pid))); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), pf.filename); err != nil {
		return err
	}
	if err := os.Chmod(pf.filename, 420); err != nil {
		return err
	}
	return nil
}

func (pf *PFile) Remove() {
	content, err := ioutil.ReadFile(pf.filename)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(string(content))
	if err != nil {
		return
	}
	if pid != pf.pid {
		return
	}
	os.Remove(pf.filename)
}

func (pf *PFile) Validate() error {
	if pf.filename == "" {
		return nil
	}

	content, err := ioutil.ReadFile(pf.filename)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			return nil
		default:
			return err
		}
	}
	pid, err := strconv.Atoi(string(content))
	if err != nil {
		return nil
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if process == nil {
		return nil
	}
	pf.pid = pid
	return nil
}
