package util

import (
	"fmt"
	"bufio"
	"os"
	"os/exec"
	"strings"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var Home = os.Getenv("HOME")

func AppendToFile (filename string, text string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

func GetApproval (prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " [y/N]")
	text, _ := reader.ReadString('\n')
	okResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	return (containsString(okResponses, strings.TrimRight(text, "\n")))
}

func BrewInstall (command string, packageName string) {
	path, err := exec.LookPath(command)
	if err != nil {
		color.Cyan("Installing %s now via brew...\n", packageName)
		ExecCmd("brew", "install", packageName)
	} else {
		color.Cyan("%s is available at %s\n", command, path)
	}
}

func BrewUninstall (command string, packageName string) {
	_, err := exec.LookPath(command)
	if err == nil {
		color.Cyan("Uninstalling %s now via brew...\n", packageName)
		ExecCmd("brew", "uninstall", packageName)
	} else {
		color.Yellow("%s doesn't seem to be available.\n", packageName)
	}
}

func BrewCaskInstall (command string, packageName string) {
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

func BrewCaskUninstall (command string, packageName string) {
	_, err := exec.LookPath(command)

	if err == nil {
		color.Cyan("Uninstalling %s now via brew cask...\n", packageName)
		ExecCmd("brew", "cask", "uninstall", packageName)
	} else {
		color.Yellow("%s doesn't seem to be available.\n", packageName)
	}
}

func AddToZshrc (replaceRegex string, text string) {
	AppendToFile(Home + "/.zshrc", text)
}

func ExecCmd(program string, args ...string) (error) {
	if viper.GetBool("debug") {
		color.Yellow("[debug] - Running '%s %s'", program, strings.Join(args, " "))
	}
	cmd := exec.Command(program, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// Using bash -c to pipe several commands, return final output as string
func ExecCmdChain(command string) (string, error){
	if viper.GetBool("debug") {
		color.Yellow("[debug] - Running commands %s through bash", command)
	}
	out, err := exec.Command("bash", "-c", command).Output()
	return string(out), err
}

func DirectoryExists(dirName string) bool {
	fileInfo, err := os.Stat(dirName)
	if err == nil && fileInfo.IsDir() {
		return true
	} else {
		return false
	}
}

