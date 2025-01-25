package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var TriggerXAsciiArt = `
	████████╗██████╗ ██╗ ██████╗  ██████╗ ███████╗██████╗ ██╗  ██╗
	╚══██╔══╝██╔══██╗██║██╔════╝ ██╔════╝ ██╔════╝██╔══██╗╚██╗██╔╝
	   ██║   ██████╔╝██║██║  ███╗██║  ███╗█████╗  ██████╔╝ ╚███╔╝
	   ██║   ██╔══██╗██║██║   ██║██║   ██║██║     ██╔══██╗ ██╔██╗
	   ██║   ██║  ██║██║╚██████╔╝╚██████╔╝███████╗██║  ██║██╔╝ ██╗
	   ╚═╝   ╚═╝  ╚═╝╚═╝ ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝
`

var TriggerXAsciiArtWithBorder = `
+------------------------------------------------------------------------+
|	████████╗██████╗ ██╗ ██████╗  ██████╗ ███████╗██████╗ ██╗  ██╗       |
|	╚══██╔══╝██╔══██╗██║██╔════╝ ██╔════╝ ██╔════╝██╔══██╗╚██╗██╔╝       |
|      ██║   ██████╔╝██║██║  ███╗██║  ███╗█████╗  ██████╔╝ ╚███╔╝        |
|	   ██║   ██╔══██╗██║██║   ██║██║   ██║██║     ██╔══██╗ ██╔██╗        |
|	   ██║   ██║  ██║██║╚██████╔╝╚██████╔╝███████╗██║  ██║██╔╝ ██╗       |
|	   ╚═╝   ╚═╝  ╚═╝╚═╝ ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝       |
+------------------------------------------------------------------------+
`

func DisplayWarningMessage(keyType, privateKey, ksPath string) error {
	cmd := exec.Command("less", "-R")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	msg := ""

	msg += fmt.Sprintln("")
	msg += fmt.Sprintln(strings.ToUpper(keyType) + " Private Key Generated Successfully 🎉")
	border := strings.Repeat("=", len(privateKey)+6)
	msg += fmt.Sprintln("\033[1m\x1b[31m" + border + "\033[0m")
	msg += fmt.Sprintln("\x1b[36m|  " + privateKey + "  |\033[0m")
	msg += fmt.Sprintln("\033[1m\x1b[31m" + border + "\033[0m")
	msg += fmt.Sprintln("")
	msg += fmt.Sprintln("\033[1m\x1b[33m🔑  WARNING: Make sure to copy this private key securely and never share it with anyone!\033[0m")
	msg += fmt.Sprintln("🔑  Keystore file will be saved at: " + ksPath)
	msg += fmt.Sprintln("Please press 'q' to exit this screen completely after confirmation.")

	if _, err = stdin.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	if err := stdin.Close(); err != nil {
		return fmt.Errorf("failed to close stdin: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for command: %w", err)
	}

	return nil
}
