package cmd

import (
	"fmt"
	"bufio"
	"os"
	"os/exec"
	"strings"
)

var home = os.Getenv("HOME")

func append(filename string, text string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
}

func get_approval (prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " [y/N]")
	text, _ := reader.ReadString('\n')
	return ("Y" == strings.TrimRight(text, "\n"))
}

func install_brew (command string, packageName string) {
	path, err := exec.LookPath(command)
	if err != nil {
		fmt.Printf("Installing %s now via brew...\n", packageName)
		exec_command("brew", "install", packageName)

	} else {
		fmt.Printf("%s is available at %s\n", command, path)
	}
}

func brew (operation string, packageName string) {
	path, err := exec.LookPath("brew")
	if err != nil {
		fmt.Printf("Installing %s now via brew...\n", packageName)
		exec_command("brew", operation, packageName)

	} else {
		fmt.Printf("%s is available at %s\n", "brew", path)
	}
}


func addToZshrc(replaceRegex string, text string) {
	append(home + "/.zshrc", text)
}

func install_brew_cask (command string, packageName string) {
	path, err := exec.LookPath(command)
	if err != nil {
		if (get_approval("Want to install " + packageName + "?") == true) {
			fmt.Printf("Installing %s now via brew cask...\n", packageName)
			exec_command("brew", "cask", "install", packageName)
		}
	} else {
		fmt.Printf("%s is available at %s\n", command, path)
	}
}


func exec_command(program string, args ...string) {
	cmd := exec.Command(program, args...)
	cmd.Stdin = os.Stdin;
	cmd.Stdout = os.Stdout;
	cmd.Stderr = os.Stderr;
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

