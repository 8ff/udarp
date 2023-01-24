package txControl

import (
	"embed"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"runtime"
	"time"
)

// Embed bins folder with rigctl binaries
//
//go:embed bin
var rigctlBins embed.FS

// Location of the temporary rigctl binary
const rigctlBinPath = "/tmp/rigctld"

var killRigctlChan = make(chan bool)

type Params struct {
	SerialPort string
	ModelId    string
	ListenPort string
}

// Function that converts Params to args for rigctl binary, skipping empty values
func ParamsToArgs(p Params) []string {
	args := []string{}

	if p.SerialPort != "" {
		args = append(args, "-r", p.SerialPort)
	}

	if p.ModelId != "" {
		args = append(args, "-m", p.ModelId)
	}

	if p.ListenPort != "" {
		args = append(args, "-t", p.ListenPort)
	}

	return args
}

func RigControlTcp(addr, port, command string) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", addr, port), time.Duration(5)*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n", command)))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Function that detects OS and architecture and stores the rigctl binary in /tmp/rigctl from rigctlBins
func fetchRigctldBinary() error {
	// Detect OS and architecture
	os := runtime.GOOS
	arch := runtime.GOARCH
	// fmt.Printf("Using rigctl-%s-%s binary\n", os, arch)

	// Find rigctl binary for OS and architecture
	var rigctlBin string
	if os == "linux" && arch == "amd64" {
		rigctlBin = "bin/rigctld-linux-amd64"
	} else if os == "linux" && arch == "arm" {
		rigctlBin = "bin/rigctld-linux-arm"
	} else if os == "linux" && arch == "arm64" {
		rigctlBin = "bin/rigctld-linux-arm64"
	} else if os == "darwin" && arch == "amd64" {
		rigctlBin = "bin/rigctld-darwin-amd64"
	} else if os == "darwin" && arch == "arm64" {
		rigctlBin = "bin/rigctld-darwin-arm64"
	} else if os == "windows" {
		rigctlBin = "bin/rigctld-windows-amd64.exe"
	} else {
		return fmt.Errorf("rigctld binary for OS %s and architecture %s not found", os, arch)
	}

	// Read rigctl binary from rigctlBins
	rigctlBinBytes, err := rigctlBins.ReadFile(rigctlBin)
	if err != nil {
		return err
	}

	// Write rigctl binary to /tmp/rigctl
	err = ioutil.WriteFile(rigctlBinPath, rigctlBinBytes, 0775)
	if err != nil {
		return err
	}

	return nil
}

// Function that runs rigctl binary with given arguments
func StartRigCtld(args []string) error {
	rigctlCmd := exec.Command(rigctlBinPath, args...)
	rigctlCmd.Stdout = nil
	rigctlCmd.Stderr = nil

	err := rigctlCmd.Start()
	if err != nil {
		return err
	}

	go func() {
		// Listen for kill
		<-killRigctlChan

		// Kill rigctl
		err := rigctlCmd.Process.Kill()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Wait for rigctl to exit
	err = rigctlCmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Stop rigctl
func StopRigCtld() {
	killRigctlChan <- true
}

// Function that runs rigctl with -V parameter to check if its working
func CheckRigCtld() error {
	// Run binary with -V parameter and check if its successful
	err := StartRigCtld([]string{"-V"})
	if err != nil {
		return fmt.Errorf("rigctld binary failed with error: %s", err)
	}

	return nil
}

func Init() error {
	err := fetchRigctldBinary()
	if err != nil {
		return err
	}

	err = CheckRigCtld()
	if err != nil {
		return err
	}

	return nil
}
