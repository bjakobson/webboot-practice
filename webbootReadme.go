package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
)

//get absolute path
func main() {
	user, _ := user.Current()
	//start under the home directory
	cmd := exec.Command("clear")
	os.Chdir(string(user.HomeDir))
	out, err := cmd.Output()

	if err != nil {
		log.Fatalf("Error changing directories %s", err)
	}
	fmt.Println(string(out))
	//list all commands
	var commands = [][]string{
		{"go", "get", "github.com/u-root/webboot"},
		{"sudo", "apt", "install", "libssl-dev"},
		{"git", "clone", "--depth", "1", "-b", "v4.12.7", "git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git", "linux"},
		{"git", "clone", "git://git.kernel.org/pub/scm/linux/kernel/git/iwlwifi/linux-firmware.git"},
		{"wget", "https://raw.githubusercontent.com/u-root/webboot/master/config-4.12.7"},
	}
	//convert [][]string to string, and execute the commands
	for _, cmd := range commands {
		c := exec.Command(cmd[0], cmd[1:]...)
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)
		}

	}
	fmt.Println("Succses")

}
u-root-practice 
# webboot-practice
