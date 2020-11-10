package ipfs

import (

	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
    "net/http"

	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

// IPFS implements communication over IPFS.
//
// We package the official IPFS daemon release for each platform into
// libipfs. This ensures the daemon operates correctly and has the
// added benefit of being easy to maintain
type IPFS struct {
	basePath   string
	daemonPath string
	daemonCmd  *exec.Cmd
}

type Error struct {
	Command string
	Message string
	Code    int
}

// New constructs a new IPFS node
func New(dataPath string) (*IPFS, error) {

	log.SetLevel(log.ErrorLevel)
	log.SetOutput(ioutil.Discard)
	os.MkdirAll(dataPath, 0744)

	binaryName := "ipfs"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	fileBytes, err := FSByte(false,
		fmt.Sprintf("/pack/%s/%s", runtime.GOOS, binaryName))
	if err != nil {
		return nil, err
	}

	daemonPath := filepath.Join(dataPath, binaryName)

	outFile, err := os.OpenFile(
		daemonPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0755)
	if err != nil {
		return nil, err
	}

	_, err = outFile.Write(fileBytes)
	if err != nil {
		return nil, err
	}
	outFile.Close()

	instance := IPFS{
		basePath:   dataPath,
		daemonPath: daemonPath,
	}

	return &instance, nil
}

// Start the IPFS node and API
func (ipfs *IPFS) Start() error {

	ipfsPath := filepath.Join(ipfs.basePath, ".scala-ipfs")
	ipfsEnv := os.Environ()
	ipfsEnv = append(ipfsEnv, fmt.Sprintf("IPFS_PATH=%s", ipfsPath))

	// Sometimes IPFS leaves an 'api' file in the path, this causes all commands
	// to fail, including starting a daemon. Let's get rid of it first
	os.Remove(filepath.Join(ipfsPath, "api"))

	// Let's first check if we have a valid IPFS repo already
	cmd := exec.Command(ipfs.daemonPath, "repo", "verify")
	cmd.Env = ipfsEnv
	op, err := cmd.CombinedOutput()
	if err != nil {
		// If we got an error that references that we need to 'ipfs init' first
		// it most likely means this is a first run
		if strings.Contains(string(op), "ipfs init") {
			cmd = exec.Command(ipfs.daemonPath, "init", "--profile", "server")
			cmd.Env = ipfsEnv
			_, err := cmd.CombinedOutput()
			if err != nil {
				// If we hit this, ipfs could not init a new repo
				return err
			}
		} else {
			// Any other error than the need for 'ipfs init' needs to be returned
			return err
		}
	}

	// Repo is good to go, we can start the daemon
	ipfs.daemonCmd = exec.Command(ipfs.daemonPath, "daemon")
	ipfs.daemonCmd.Env = ipfsEnv
	err = ipfs.daemonCmd.Start()
	if err != nil {
		return err
	}

	go func() {
		err = ipfs.daemonCmd.Wait()
		fmt.Println("IPFS daemon completed with exit:", err)
	}()
	// Give the daemon some time to start up
	// TODO: I don't have a simple way to check this... yet
	time.Sleep(time.Second * 5)
	return nil
}

// Get an object from IPFS and return it as bytes
func (ipfs *IPFS) Get(hash string) ([]byte, error) {

	downloadPath := filepath.Join(ipfs.basePath, hash)
	sh := shell.NewShell("localhost:5001")
	err := sh.Get(hash, downloadPath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(downloadPath)
}

// Resolve an IPNS name to IPFS hash
func (ipfs *IPFS) Resolve(peerid string) (string, error) {

	sh := shell.NewShell("localhost:5001")
	resp, err := sh.Resolve(peerid)
	if err != nil {
		return "", err
	}
	return resp, nil
}

// Add a directory to IPFS
func (ipfs *IPFS) AddDirectory(directory string) (string, error) {

	sh := shell.NewShell("localhost:5001")
	resp, err := sh.AddDir(directory)
	if err != nil {
		return "", err
	}
	return resp, nil
}

// Add bootstrap nodes to IPFS
func (ipfs *IPFS) BootstrapAdd(peers []string) (string, error) {

	sh := shell.NewShell("localhost:5001")
	resp, err := sh.BootstrapAdd(peers)
	if err != nil {
		return "", err
	}

	resp2 := strings.Join(resp, ",")
	return resp2, nil
}


// Get PeerID
func (ipfs *IPFS) GetPeerID() (string, error) {
    
    sh:= shell.NewShell("localhost:5001")
    resp, err:= sh.ID();
    if err != nil {
        return "", err
    }
    return resp.ID, nil
}

// Publish a content hash to IPNS
func (ipfs *IPFS) PublishName(contentHash string) (string, error) {
    sh:= shell.NewShell("localhost:5001")
    err:= sh.Publish("", contentHash);
    if err != nil {
        return "",err
    }
    return "Successfully published to IPNS",nil
}


// Stop the IPFS node
func (ipfs *IPFS) Stop() (error) {

    client := &http.Client{}
    req, err := http.NewRequest("POST", "http://localhost:5001/api/v0/shutdown", nil)
    if err != nil {
        return err
    }

    resp, err := client.Do(req)
    _ = resp
    if err != nil {
        return err
    }

    return nil
}
