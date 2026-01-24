package injector

import (
	"os"
	"os/exec"
)

func Run(command string, args []string, secrets map[string]string) error {
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
