package kube

import (
	"os"
	"os/exec"
)

// Exec runs kubectl exec to attach an interactive terminal to the specified
// pod. If command is empty, it defaults to /bin/sh.
func Exec(namespace, podName string, command []string) error {
	if len(command) == 0 {
		command = []string{"/bin/sh"}
	}

	args := []string{"exec", "-it", "-n", namespace, podName, "--"}
	args = append(args, command...)

	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
