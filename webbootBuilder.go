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
	devB1      = "/dev/sdb1"
	devBSTATUS = false

	devC       = "/dev/sdc"
	devCSTATUS = false
	devC1      = "/dev/sdc1"

	devD       = "/dev/sdd"
	devD1      = "/dev/sdD1"
	devDSTATUS = false

	noUSB = true
)

// check what device exists

// this way we arent trying every possible device in future uses

// typical error message would return "permission denied"
//if it find /dev/sdX, so i can render the error message as a boolean
func Device() bool {

	b, err := os.Open("/dev/sdb")
	if strings.Contains(err.Error(), "permission denied") {
		devBSTATUS = true
		noUSB = false
	} else if !strings.Contains(err.Error(), "no such file or directory") {
		devBSTATUS = false
	}

	c, err := os.Open("/dev/sdc")
	if strings.Contains(err.Error(), "permission denied") {

		devCSTATUS = true
		noUSB = false
	} else if !strings.Contains(err.Error(), "no such file or directory") {
		devCSTATUS = false
	}

	d, err := os.Open("/dev/sdd")
	if strings.Contains(err.Error(), "permission denied") {

		devDSTATUS = true
		noUSB = false
	} else if !strings.Contains(err.Error(), "no such file or directory") {
		devDSTATUS = false
	}

	if noUSB == true {
		log.Fatal("No USB deteced, please insert a USB stick")
	}
	_ = b
	_ = c
	_ = d

	return devBSTATUS
	return devCSTATUS
	return devDSTATUS
}

func main() {
	//executeCommands()
	Webboot()

	Device()

	fmt.Println("devB:", devBSTATUS, "\n devC:", devCSTATUS, "\n devD:", devDSTATUS)
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
var name1 string
var mkp string
var of string

func Init() {
	if devBSTATUS == true {
		name = "/dev/sdb"
		name1 = devB1
		mkp = "echo 'type=83' | sudo sfdisk /dev/sdb \n sudo mkfs -t vfat /dev/sdb1"
		of = "of=/dev/sdb"
	} else if devCSTATUS == true {
		name = "/dev/sdc"
		name1 = devC1
		mkp = "echo 'type=83' | sudo sfdisk /dev/sdc \n sudo mkfs -t vfat /dev/sdc1"
		of = "of=/dev/sdc"

	} else if devDSTATUS == true {
		name = "/dev/sdd"
		name1 = devD1
		of = "of=/dev/sdd"

		mkp = "echo 'type=83' | sudo sfdisk /dev/sdd \n sudo mkfs -t vfat /dev/sdd1"

	} else {
		fmt.Println()
	}
}

//building usb
func DeletePartition(path error) {
	var command = [][]string{

		{"sudo", "umount", name},

		//{"sudo", "fdisk", "/dev/sdb"},

		//{"echo", "d"},
		//{"echo", "1"},
		//{"echo", "w"},
	}

	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			fmt.Printf("\n\n\n%s failed: ignored %v", cmd, err)
			fmt.Println("\n\n\n")
		}

	}
	//os.Exit(0)
}

func MakePartition(path error) {
	f, err := os.Create("makePartition.sh")
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.WriteString(mkp)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l)
	f.Close()
	var command = [][]string{
		{"pwd"},
		{"sudo", "bash", "./makePartition.sh"},
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

func DownloadFile(path error) {
	var command = [][]string{
		{"pwd"},
		{"sudo", "apt-get", "install", "mtools"},
		{"sudo", "apt-get", "install", "libc6-i386"},
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
		{"sudo", "mount", name1, "USB"},
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
		{"sudo", "./syslinux", "-i", name1},
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
		{"sudo", "dd", "bs=440", "count=1", "conv=notrunc", "if=mbr.bin", of},
		{"sudo", "parted", name, "set", "1", "boot", "on"},
	}
	for _, cmd := range command {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("%s failed: %v", cmd, err)

		}

	}
}
