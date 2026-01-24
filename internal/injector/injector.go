package injector

import (
	"os"
	"os/exec"
)

type Runner interface {
	Run(command string, args []string, secrets map[string]string) error
}

type runner struct{}

func NewRunner() Runner {
	return &runner{}
}

func (r *runner) Run(command string, args []string, secrets map[string]string) error {
	cmd := exec.Command(command, args...)

	env := os.Environ()
	for k, v := range secrets {
		env = append(env, k+"="+v)
	}

	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func Run(command string, args []string, secrets map[string]string) error {
	r := NewRunner()
	return r.Run(command, args, secrets)
}
