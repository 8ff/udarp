package txControl

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/8ff/udarp/pkg/misc"
)

// Embed bins folder with rigctl binaries
//
//go:embed bin
var rigctlBins embed.FS

type TxControl struct {
	Params           Params
	RigctlBinPath    string
	StopRigCtrldChan chan bool
}

type Params struct {
	SerialPort string
	BaudRate   string
	ModelId    string
	ListenPort string
	ListenAddr string
}

func New(params Params) (*TxControl, error) {
	// Go over all the params and make sure they are set
	if params.SerialPort == "" {
		return nil, fmt.Errorf("serial port not set")
	}

	if params.BaudRate == "" {
		return nil, fmt.Errorf("baud rate not set")
	}

	if params.ModelId == "" {
		return nil, fmt.Errorf("model id not set")
	}

	if params.ListenAddr == "" {
		return nil, fmt.Errorf("listen address not set")
	}

	if params.ListenPort == "" {
		return nil, fmt.Errorf("listen port not set")
	}

	t := &TxControl{
		Params: params,
	}

	// Randomly generate name for rigctl binary using secure random number generator
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Use this path to store temporary rigctl binary
	t.RigctlBinPath = fmt.Sprintf("/tmp/rigctl-%s", hex.EncodeToString(randomBytes))

	// Fetch rigctl binary from embedded files
	err = t.fetchRigctldBinary()
	if err != nil {
		return nil, err
	}

	// Check to make sure that binary is working
	err = t.startRigCtld([]string{"-V"})
	if err != nil {
		return nil, fmt.Errorf("rigctld binary failed with error: %s", err)
	}

	return t, nil
}

func (t *TxControl) startRigCtld(args []string) error {
	rigctlCmd := exec.Command(t.RigctlBinPath, args...)
	rigctlCmd.Stdout = nil
	rigctlCmd.Stderr = nil

	err := rigctlCmd.Start()
	if err != nil {
		return err
	}

	go func() {
		// Listen for kill
		<-t.StopRigCtrldChan

		// Kill rigctl
		err := rigctlCmd.Process.Kill()
		if err != nil {
			misc.Log("error", fmt.Sprintf("error killing rigctl: %s", err))
		}
	}()

	// Wait for rigctl to exit
	err = rigctlCmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Convert Params to flags
func (t *TxControl) Start() error {
	// Start rigctl
	err := t.startRigCtld([]string{
		"-r", t.Params.SerialPort,
		"-m", t.Params.ModelId,
		"-t", t.Params.ListenPort,
		"-s", t.Params.BaudRate,
		"-T", t.Params.ListenAddr,
	})
	if err != nil {
		return fmt.Errorf("error starting rigctl: %s", err)
	}

	return nil
}

// Func that kills rigctl
func (t *TxControl) Stop() {
	// Kill rigctl
	t.StopRigCtrldChan <- true

	// Remove rigctl binary
	exec.Command("rm", t.RigctlBinPath).Run()
}

func (t *TxControl) TcpCommand(addr, port, command string) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", addr, port), time.Duration(5)*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n", command)))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	readBytes, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:readBytes], nil
}

// TX turns on the transmitter
func (t *TxControl) TX() error {
	// Verify that the radio is in TX mode by sending 't' and checking if output is 1
	for retry := 0; retry < 3; retry++ {
		_, err := t.TcpCommand(t.Params.ListenAddr, t.Params.ListenPort, "T 1")
		if err != nil {
			return fmt.Errorf("error sending TX command: %s", err)
		}

		buf, err := t.TcpCommand(t.Params.ListenAddr, t.Params.ListenPort, "t")
		if err != nil {
			return fmt.Errorf("error sending get TX status command: %s", err)
		}

		// Strip newline
		buf = buf[:len(buf)-1]

		if string(buf) == "1" {
			return nil
		}

		time.Sleep(time.Duration(1) * time.Second)
	}

	return fmt.Errorf("unabe to put radio in TX mode after 3 tries")
}

// RX turns off the transmitter
func (t *TxControl) RX() error {
	// Verify that the radio is in RX mode by sending 't' and checking if output is 0
	for retry := 0; retry < 3; retry++ {
		_, err := t.TcpCommand(t.Params.ListenAddr, t.Params.ListenPort, "T 0")
		if err != nil {
			return fmt.Errorf("error sending RX command: %s", err)
		}

		buf, err := t.TcpCommand(t.Params.ListenAddr, t.Params.ListenPort, "t")
		if err != nil {
			return fmt.Errorf("error sending get TX status command: %s", err)
		}

		// Strip newline
		buf = buf[:len(buf)-1]

		if string(buf) == "0" {
			return nil
		}

		time.Sleep(time.Duration(1) * time.Second)
	}

	return fmt.Errorf("unable to put radio into RX mode after 3 tries")
}

// Get the current frequency as string
func (t *TxControl) GetFrequencyString() (string, error) {
	buf, err := t.TcpCommand(t.Params.ListenAddr, t.Params.ListenPort, "f")
	if err != nil {
		return "", fmt.Errorf("error sending get frequency command: %s", err)
	}

	return string(buf), nil
}

// Get the current frequency as float64
func (t *TxControl) GetFrequency() (float64, error) {
	buf, err := t.TcpCommand(t.Params.ListenAddr, t.Params.ListenPort, "f")
	if err != nil {
		return 0, fmt.Errorf("error sending get frequency command: %s", err)
	}

	freq, err := strconv.ParseFloat(string(buf), 64)
	if err != nil {
		return 0, fmt.Errorf("error converting frequency to float: %s", err)
	}

	return freq, nil
}

// Set frequency
func (t *TxControl) SetFrequency(freq float64) error {
	_, err := t.TcpCommand(t.Params.ListenAddr, t.Params.ListenPort, fmt.Sprintf("F %f", freq))
	if err != nil {
		return fmt.Errorf("error sending set frequency command: %s", err)
	}

	return nil
}

// Function that detects OS and architecture and stores the rigctl binary in /tmp/rigctl from rigctlBins
func (t *TxControl) fetchRigctldBinary() error {
	// Detect OS and architecture
	systemOs := runtime.GOOS
	arch := runtime.GOARCH
	// fmt.Printf("Using rigctl-%s-%s binary\n", os, arch)

	// Find rigctl binary for OS and architecture
	var rigctlBin string
	if systemOs == "linux" && arch == "amd64" {
		rigctlBin = "bin/rigctld-linux-amd64"
	} else if systemOs == "linux" && arch == "arm" {
		rigctlBin = "bin/rigctld-linux-arm"
	} else if systemOs == "linux" && arch == "arm64" {
		rigctlBin = "bin/rigctld-linux-arm64"
	} else if systemOs == "darwin" && arch == "amd64" {
		rigctlBin = "bin/rigctld-darwin-amd64"
	} else if systemOs == "darwin" && arch == "arm64" {
		rigctlBin = "bin/rigctld-darwin-arm64"
	} else if systemOs == "windows" {
		rigctlBin = "bin/rigctld-windows-amd64.exe"
	} else {
		return fmt.Errorf("rigctld binary for OS %s and architecture %s not found", systemOs, arch)
	}

	// Read rigctl binary from rigctlBins
	rigctlBinBytes, err := rigctlBins.ReadFile(rigctlBin)
	if err != nil {
		return err
	}

	// Write rigctl binary to /tmp/rigctl
	err = os.WriteFile(t.RigctlBinPath, rigctlBinBytes, 0775)
	if err != nil {
		return err
	}

	return nil
}
