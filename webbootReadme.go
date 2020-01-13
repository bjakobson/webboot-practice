package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
)

func executeCommands() {

	var (
		user, _   = user.Current()
		home      = string(user.HomeDir)
		homeLinux = path.Join(string(user.HomeDir), "/linux")
	)

	//under home dir
	os.Chdir(home)
	var commandsHome = [][]string{
		{"go", "get", "github.com/u-root/webboot"},
		{"sudo", "apt", "install", "libssl-dev", "build-essential"},
		{"git", "clone", "--depth", "1", "-b", "v4.12.7", "git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git", "linux"},
		{"git", "clone", "git://git.kernel.org/pub/scm/linux/kernel/git/iwlwifi/linux-firmware.git"},
	}

	//under 'linux'
	var commandsLinux = [][]string{
		{"wget", "https://raw.githubusercontent.com/u-root/webboot/master/config-4.12.7"},
		{"cp", "config-4.12.7", ".config"},
		{"make", "bzImage"},
	}

	fmt.Println("Downloading and storing files...")

	//execute home
	for _, cmd := range commandsHome {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)
		}

	}

	os.Chdir(homeLinux)
	for _, cmd := range commandsLinux {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)
		}

	}

}

func main() {
	executeCommands()
	fmt.Println("Succsesfully downloaded and stored all files!")

}
