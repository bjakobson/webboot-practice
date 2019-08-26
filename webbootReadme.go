package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
)

func main() {

	//get users home directory/name, and start all downloads under home/usr
	user, _ := user.Current()
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
	}
	//convert [][]string to string, and execute the commands
	for _, cmd := range commands {
		c := exec.Command(cmd[0], cmd[1:]...)
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)
		}

	}

	//set download location of config to linux
	homeLinux := path.Join(string(user.HomeDir), "/linux")
	getConfig := exec.Command("wget", "https://raw.githubusercontent.com/u-root/webboot/master/config-4.12.7")
	os.Chdir(homeLinux)
	Config, err := getConfig.Output()
	fmt.Println(string(Config))
	if err != nil {
		log.Fatalf("Error getting config file %s", err)
	}
	fmt.Println(string(out))

	getbzImage := exec.Command("wget", "https://github.com/bjakobson/webboot-practice/blob/master/bzImage")
	os.Chdir(homeLinux)
	bzImage, err := getbzImage.Output()
	fmt.Println(string(bzImage))
	if err != nil {
		log.Fatalf("Error getting bzImage file %s", err)
	}

	file := exec.Command("go", "run", ".")
	runFiles, _ := file.CombinedOutput()

	fmt.Println(runFiles)

	fmt.Println("Succsesfully downloaded and stored all files!")

}
