package vms

import (
	"fmt"
	"github.com/mhewedy/vermin/cmd"
	"github.com/mhewedy/vermin/cmd/scp"
	"github.com/mhewedy/vermin/cmd/ssh"
	"github.com/mhewedy/vermin/db"
	"github.com/mhewedy/vermin/ip"
	"os"
)

func Tag(vmName string, tag string) error {
	if err := checkRunningVM(vmName); err != nil {
		return err
	}

	return appendToFile(db.GetVMPath(vmName)+"/"+db.Tags, []byte(tag+"\n"), 0775)
}

func Start(vmName string) error {
	if err := checkVM(vmName); err != nil {
		return err
	}

	fmt.Println("Starting", vmName, "...")
	if _, err := cmd.Execute("vboxmanage", "startvm", vmName, "--type", "headless"); err != nil {
		return err
	}
	return nil
}

func Stop(vmName string) error {
	if err := checkVM(vmName); err != nil {
		return err
	}

	fmt.Println("Stopping", vmName, "...")
	if _, err := cmd.Execute("vboxmanage", "controlvm", vmName, "poweroff"); err != nil {
		return err
	}
	return nil
}

func SecureShell(vmName string, command string) error {
	if err := checkRunningVM(vmName); err != nil {
		return err
	}

	if err := ssh.EstablishConn(vmName); err != nil {
		return err
	}
	if len(command) == 0 {
		return ssh.OpenTerminal(vmName)
	} else {
		return ssh.ExecuteI(vmName, command)
	}
}

func Remove(vmName string) error {
	if err := checkVM(vmName); err != nil {
		return err
	}
	_ = Stop(vmName)
	fmt.Println("Removing", vmName, "...")
	if _, err := cmd.Execute("vboxmanage", "unregistervm", vmName, "--delete"); err != nil {
		return err
	}
	return os.RemoveAll(db.GetVMPath(vmName))
}

func PortForward(vmName string, ports string) error {
	if err := checkRunningVM(vmName); err != nil {
		return err
	}

	a, err := getPortForwardArgs(ports)
	if err != nil {
		return err
	}
	if err := ssh.EstablishConn(vmName); err != nil {
		return err
	}
	fmt.Println("Connected. Press CTRL+C anytime to stop")
	if err := ssh.ExecuteIArgs(vmName, append(a, "-N")...); err != nil {
		return err
	}

	return nil
}

func CopyFiles(vmName string, file string, toVM bool) error {
	if err := checkRunningVM(vmName); err != nil {
		return err
	}

	if toVM {
		return scp.CopyToVMHomeDir(vmName, file)
	} else {
		return scp.CopyToLocalCWD(vmName, file)
	}
}

func IP(vmName string, purge bool) (string, error) {
	if err := checkRunningVM(vmName); err != nil {
		return "", err
	}

	return ip.Find(vmName, purge)
}

func checkRunningVM(vmName string) error {
	list, err := List(false)
	if err != nil {
		return err
	}
	if !contains(list, vmName) {
		return fmt.Errorf("%s not running.\nUse the command 'vermin ps' to list running VMs", vmName)
	}
	return nil
}

func checkVM(vmName string) error {
	list, err := List(true)
	if err != nil {
		return err
	}
	if !contains(list, vmName) {
		return fmt.Errorf("%s not found.\nUse the command 'vermin ps -a' to list VMs", vmName)
	}
	return nil
}

func contains(a []string, s string) bool {
	for i := range a {
		if a[i] == s {
			return true
		}
	}
	return false
}