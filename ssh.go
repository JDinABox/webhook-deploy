package webhookdeploy

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/skeema/knownhosts"
	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	Client *ssh.Client
	Config *Remote
}

func NewSSHClient(config *Remote, knownHost string) (*SSHClient, error) {
	key, err := os.ReadFile(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	f, err := os.OpenFile(knownHost, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	kh, err := knownhosts.NewDB(knownHost)
	if err != nil {
		log.Fatal("Failed to read known_hosts: ", err)
	}

	// Create a custom permissive hostkey callback which still errors on hosts
	// with changed keys, but allows unknown hosts and adds them to known_hosts
	cb := ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		innerCallback := kh.HostKeyCallback()
		err := innerCallback(hostname, remote, key)
		if knownhosts.IsHostKeyChanged(err) {
			return fmt.Errorf("REMOTE HOST IDENTIFICATION HAS CHANGED for host %s! This may indicate a MitM attack.", hostname)
		} else if knownhosts.IsHostUnknown(err) {
			f, fErr := os.OpenFile(knownHost, os.O_APPEND|os.O_WRONLY, 0600)
			if fErr != nil {
				requestLogger.Logf("Failed to open known_hosts: %v\n", fErr)
				return nil
			}
			defer f.Close()

			fErr = knownhosts.WriteKnownHost(f, hostname, remote, key)
			if fErr != nil {
				requestLogger.Logf("Failed to add host %s to known_hosts: %v\n", hostname, fErr)
				return nil
			}

			log.Printf("Added host %s to known_hosts\n", hostname)
			return nil // permit previously-unknown hosts (warning: may be insecure)
		}
		return err
	})

	algorithms := ssh.SupportedAlgorithms()

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback:   cb,
		HostKeyAlgorithms: algorithms.HostKeys,
		Timeout:           10 * time.Second,
	}

	addr := net.JoinHostPort(config.ServerIP, "22")
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &SSHClient{Client: client, Config: config}, nil
}

func (sc *SSHClient) RunCommand(command string) (string, string, error) {
	session, err := sc.Client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("unable to create stdout pipe: %w", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("unable to create stderr pipe: %w", err)
	}

	err = session.Start(command)
	if err != nil {
		return "", "", fmt.Errorf("failed to start command: %w", err)
	}

	stdoutStr := ""
	stderrStr := ""

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		outBytes, readErr := io.ReadAll(stdout)
		if readErr != nil {
			log.Printf("[%s] Error reading stdout: %v", sc.Config.ServerIP, readErr)
		}
		stdoutStr = string(outBytes)
	}()

	go func() {
		defer wg.Done()
		errBytes, readErr := io.ReadAll(stderr)
		if readErr != nil {
			log.Printf("[%s] Error reading stderr: %v", sc.Config.ServerIP, readErr)
		}
		stderrStr = string(errBytes)
	}()

	wg.Wait()

	err = session.Wait()
	if err != nil {
		return stdoutStr, stderrStr, err
	}

	return stdoutStr, stderrStr, nil
}
