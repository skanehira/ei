package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/containerd/console"
	"github.com/kr/pty"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var rootCmd = &cobra.Command{
	Use: "ei",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func exitError(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func Execute() {
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		current := console.Current()
		current.DisableEcho()
		defer current.Reset()

		if err := current.SetRaw(); err != nil {
			exitError(err)
		}

		term := terminal.NewTerminal(current, "")

		c := exec.Command(args[0], args[1:]...)
		ptmx, err := pty.Start(c)
		if err != nil {
			exitError(err)
		}

		go func() {
			_, _ = io.Copy(os.Stdout, ptmx)
		}()

		for {
			line, err := term.ReadLine()
			if err != nil {
				break
			}
			_, err = ptmx.WriteString(line + "\n")
			if err != nil {
				exitError(err)
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		exitError(err)
	}
}
