package util

import (
	"fmt"
	"bufio"
	"os"
	"os/exec"
	"strings"
	"github.com/fatih/color"
	"io/ioutil"
)

var Home = os.Getenv("HOME")

func AppendToFile(filename string, text string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
}

func GetApproval (prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " [y/N]")
	text, _ := reader.ReadString('\n')
	return ("Y" == strings.TrimRight(text, "\n"))
}

func Brew (operation string, packageName string) {
	path, err := exec.LookPath("brew")
	if err != nil {
		fmt.Printf("Installing %s now via brew...\n", packageName)
		ExecCmd("brew", operation, packageName)
	} else {
		fmt.Printf("%s is available at %s\n", "brew", path)
	}
}

func InstallBrew (command string, packageName string) {
	path, err := exec.LookPath(command)
	if err != nil {
		color.Cyan("Installing %s now via brew...\n", packageName)
		ExecCmd("brew", "install", packageName)
	} else {
		color.Cyan("%s is available at %s\n", command, path)
	}
}

func InstallBrewCask (command string, packageName string) {
	path, err := exec.LookPath(command)
	if err != nil {
		if (GetApproval("Want to install " + packageName + "?") == true) {
			fmt.Printf("Installing %s now via brew cask...\n", packageName)
			ExecCmd("brew", "cask", "install", packageName)
		}
	} else {
		color.Cyan("%s is available at %s\n", command, path)
	}
}

func AddToZshrc(replaceRegex string, text string) {
	AppendToFile(Home + "/.zshrc", text)
}

func ExecCmd(program string, args ...string) {
	// Push stdout into byte buffer instead of to user's shell
	cmd := exec.Command(program, args...)
	cmd.Stdin = os.Stdin;
	cmd.Stdout = ioutil.Discard;
	cmd.Stderr = os.Stderr;
	err := cmd.Run()
	if err != nil {
		color.Red("%v\n", err)
	}
}

