package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

var (
	devB       = "/dev/sdb"
	devBSTATUS = false
	devC       = "/dev/sdc"
	devCSTATUS = false
	devD       = "/dev/sdd"
	devDSTATUS = false
)

func Device() bool {
	if fileExists(devB) {
		fmt.Println("Found /dev/sdb")
		devBSTATUS = true
	} else if fileExists(devC) {
		fmt.Println("found /dev/sdc")
		devCSTATUS = true
	} else if fileExists(devD) {
		fmt.Println("found /dev/sdd")
		devDSTATUS = true

	} else {
		log.Fatalf("Please insert a USB")
	}

	return devBSTATUS
	return devCSTATUS
	return devDSTATUS
}

func main() {
	executeCommands()
	Webboot()

	Device()
	//making sure syslinux isn't already downloaded
	if fileExists("/home/brandonjakobson/Downloads/syslinux-6.04-pre1.tar.gz") {
		fmt.Println("Syslinux exists, not downloading")
	} else {
		DownloadFile(os.Chdir(filepath.Join(homeDir, "/Downloads")))
	}

	Init()

	DeletePartition(os.Chdir(homeDir))

	MakePartition(os.Chdir(homeDir))

	Mount(os.Chdir(homeDir))

	activate(os.Chdir(bioslinux))

	syslinux(os.Chdir(homeDir))

	BootBuild(os.Chdir(usb))

	kernelInitramfs(os.Chdir(boot))

	dd(os.Chdir(biosmbr))

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var (
	usr, _  = user.Current()
	homeDir = usr.HomeDir

	bioslinux = filepath.Join(homeDir, "/Downloads/syslinux-6.04-pre1/bios/linux")
	biosmbr   = filepath.Join(homeDir, "/Downloads/syslinux-6.04-pre1/bios/mbr")

	libutil        = filepath.Join(homeDir, "/Downloads/syslinux-6.04-pre1/bios/com32/libutil/libutil.c32")
	libcom32       = filepath.Join(homeDir, "/Downloads/syslinux-6.04-pre1/bios/com32/lib/libcom32.c32")
	vesamenu       = filepath.Join(homeDir, "/Downloads/syslinux-6.04-pre1/bios/com32/menu/vesamenu.c32")
	syslinuxConfig = filepath.Join(homeDir, "/syslinux.cfg")
	usb            = filepath.Join(homeDir, "/USB")
	boot           = filepath.Join(homeDir, "/USB/boot")
	kernel         = filepath.Join(homeDir, "/linux/arch/x86/boot/bzImage")
	initramfs      = "/tmp/initramfs.linux_amd64.cpio"
)

//download webboot files
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
		{"go", "get", "github.com/vishvananda/netlink"},
		{"sudo", "apt", "install", "libssl-dev", "build-essential"},
		{"git", "clone", "--depth", "1", "-b", "v4.12.7",
			"git://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git",
			"linux"},
		{"git", "clone",
			"git://git.kernel.org/pub/scm/linux/kernel/git/iwlwifi/linux-firmware.git"},
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

//buildimage.go with my own changes
var (
	debug = func(string, ...interface{}) {}

	verbose = flag.Bool("v", true, "verbose debugging output")
	uroot   = flag.String("u", "-build=bb", "options for u-root")
	cmds    = flag.String("c", "core", "u-root commands to build into the image")
	wcmds   = flag.String("w", "github.com/u-root/webboot/webboot/.",
		"webboot commands to build into the image")
	ncmds = flag.String("n", "github.com/u-root/NiChrome/cmds/wifi",
		"NiChrome commands to build into the image")
)

func init() {
	flag.Parse()
	if *verbose {
		debug = log.Printf
	}
}

// This function is a bit nasty but we'll need it until we can extend
// u-root a bit. Consider it a hack to get us ready for OSFC.
// the Must means it has to succeed or we die.
func extraBinMust(n string) string {
	p, err := exec.LookPath(n)
	if err != nil {
		log.Fatalf("extraMustBin(%q): %v; you may need to run sudo apt installwireless-tools wpasupplicant", n, err)
	}
	return p
}

func Webboot() {
	usr, _ := user.Current()
	var homeDir = string(usr.HomeDir)
	var WheretoBuild = filepath.Join(homeDir,
		"/linux/arch/x86/boot/bzImage:bzImage")

	var commands = [][]string{
		{"date"},
		{"go", "get", "-u", "github.com/u-root/u-root"},
		{"go", "get", "-d", "-v", "-u", "github.com/u-root/NiChrome/..."},
		append(append([]string{"go", "run", "github.com/u-root/u-root/.",
			"-build=bb", "-files", WheretoBuild,
			"-files", extraBinMust("iwconfig"),
			"-files", extraBinMust("iwlist"),
			"-files", extraBinMust("wpa_supplicant"),
			"-files", extraBinMust("wpa_action"),
			"-files", extraBinMust("wpa_cli"),
			"-files", extraBinMust("wpa_passphrase"),
		}, strings.Fields(*uroot)...), *cmds, *wcmds, *ncmds),
	}
	for _, cmd := range commands {
		debug("Run %v", cmd)
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)
		}
	}
	debug("done")
}

var name string

func Init() {
	if devBSTATUS == true {
		name = "/dev/sdb"
	} else if devCSTATUS == true {
		name = "/dev/sdc"
	} else if devDSTATUS == true {
		name = "/dev/sdd"
	} else {
		fmt.Println()
	}
	fmt.Println(name)
}

//building usb
func DeletePartition(path error) {
	var command = [][]string{

		{"sudo", "umount", name},

		{"sudo", "fdisk", "/dev/sdc"},

		{"echo", "d"},
		//{"echo", "1"},
		{"echo", "w"},
	}

	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			fmt.Printf("\n\n\n%s failed: ignored %v", cmd, err)
			fmt.Println("\n\n\n")
		}

	}

}

func MakePartition(path error) {
	var command = [][]string{
		{"pwd"},
		//{"sudo", "dd", "if=/dev/zero", "of=/dev/sdb", "bs=4k"},
		{"sudo", "fdisk", name},

		{"echo", "n"},
		{"echo", "p"},
		{"echo", "w"},

		{"sudo", "mkfs.vfat", "-I", name},

		{"sudo", "mount", "-o", "remount,rw", name, usb},
	}

	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			fmt.Printf("\n\n\n%s failed: ignored %v", cmd, err)
			fmt.Println("\n\n\n")
		}

	}

	os.Exit(0)
}

func DownloadFile(path error) {
	var command = [][]string{
		{"pwd"},
		{"sudo", "apt-get", "install", "mtools"},
		{"sudo", "apt-get", "install", "libc6-i386"},
		{"echo", "Y"},
		{"sudo", "apt-get", "install", "wget"},
		{"echo", "Y"},
		{"wget", "https://mirrors.edge.kernel.org/pub/linux/utils/boot/syslinux/Testing/6.04/syslinux-6.04-pre1.tar.gz"},
		{"sudo", "tar", "-xzf", "syslinux-6.04-pre1.tar.gz"},
	}
	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)

		}

	}
}

func Mount(path error) {
	var command = [][]string{
		{"sudo", "mkdir", "USB"},
		{"sudo", "mount", name, "USB"},
	}

	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			fmt.Printf("\n\n\n%s failed: ignored", cmd)
			fmt.Println("\n\n\n")
		}

	}
}

func activate(path error) {
	var command = [][]string{
		{"sudo", "./syslinux", "-i", name},
	}

	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			fmt.Printf("\n\n\n%s failed: ignored", cmd)
			fmt.Println("\n\n\n")
		}

	}
}

func syslinux(path error) {
	var command = [][]string{
		{"sudo", "cp", "-R", libutil, libcom32, vesamenu, usb},
		{"wget", "-O", "syslinux.cfg", "https://pastebin.com/raw/mm2eVh6Y"},
		{"sudo", "cp", "-R", syslinuxConfig, usb},
	}
	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)

		}

	}
}

func BootBuild(path error) {
	var command = [][]string{
		{"sudo", "mkdir", "boot"},
	}

	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {

			fmt.Printf("\n\n\n%s failed: ignored", cmd)
			fmt.Println("\n\n\n")
		}
	}
}

func kernelInitramfs(path error) {
	var command = [][]string{
		{"sudo", "cp", "-R", kernel, initramfs, boot},
	}

	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)
		}

	}

}

func dd(path error) {
	var command = [][]string{
		{"pwd"},
		{"sudo", "dd", "bs=440", "count=1", "conv=notrunc", "if=mbr.bin", "of=/dev/sdb"},
	}
	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)

		}

	}
}
