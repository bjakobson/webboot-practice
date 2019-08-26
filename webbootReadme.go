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
		user, _ = user.Current()
		cmd     = exec.Command("clear")
		out, _  = cmd.Output()

		homeLinux = path.Join(string(user.HomeDir), "/linux")

		buildecc = exec.Command("sudo", "apt-get", "install", "build-essential")
	)

	//get users home directory/name, and start all downloads under home/user
	os.Chdir(string(user.HomeDir))
	fmt.Println(string(out))

	//list all commands
	var commands = [][]string{
		{"go", "get", "github.com/u-root/webboot"},
		{"sudo", "apt", "install", "libssl-dev"},
		{"git", "clone", "--depth", "1", "-b", "v4.12.7", "git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git", "linux"},
		{"git", "clone", "git://git.kernel.org/pub/scm/linux/kernel/git/iwlwifi/linux-firmware.git"},
	}
	//convert [][]string to string, and execute the commands
	for _, cmd := range commands {
		c := exec.Command(cmd[0], cmd[1:]...)
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)
		}

	}

	//set download location of config to linux
	os.Chdir(homeLinux)
	getConfig := exec.Command("wget", "https://raw.githubusercontent.com/u-root/webboot/master/config-4.12.7")
	Config, _ := getConfig.Output()
	fmt.Println(string(out), string(Config))

	buildecc.CombinedOutput()

}

func main() {
	executeCommands()
	fmt.Println("Succsesfully downloaded and stored all files!")

}
