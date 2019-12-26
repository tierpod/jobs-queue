package worker

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/shlex"
	"github.com/tierpod/jobs-queue/app/config"
)

// Command contains executable and arguments.
type Command struct {
	Executable string
	Args       []string
}

// NewCommand creates new Command from `s`.
func NewCommand(s string, cfg *config.Config) (Command, error) {
	ss, err := shlex.Split(s)
	if err != nil {
		return Command{}, err
	}

	executable := ss[0]
	args := ss[1:]

	if ok := stringSliceContains(cfg.Jobs, executable); !ok {
		return Command{}, fmt.Errorf("job executable '%v' not configured", executable)
	}

	return Command{
		Executable: executable,
		Args:       args,
	}, nil
}

// Key is the key for using in cache.
func (c Command) Key() string {
	return c.String()
}

func (c Command) String() string {
	return fmt.Sprintf("%v %v", c.Executable, strings.Join(c.Args, " "))
}

// Exec runs command and returns stdout, stderr buffers.
func (c Command) Exec() (bytes.Buffer, bytes.Buffer, error) {
	cmd := exec.Command(c.Executable, c.Args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, err
	}

	return stdout, stderr, nil
}

func stringSliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}

	return false
}
