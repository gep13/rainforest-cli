package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/rainforestapp/rainforest-cli/rainforest"
	"github.com/urfave/cli"

	"github.com/localtunnel/go-localtunnel"
)

func GetTunnel(port int, host string) (*localtunnel.LocalTunnel, error) {
	tunnel, err := localtunnel.New(port, host, localtunnel.Options{
		BaseURL: "http://rainforest.run",
	})

	if err != nil {
		fmt.Println("Error: %v", err)
	}

	return tunnel, err
}

type localRunner struct {
	client runnerAPI
}

func startLocalRun(c cliContext) error {
	r := newLocalRunner()
	return r.startRun(c)
}

func startLocalTestEditing(c cliContext) error {
	r := newLocalRunner()
	return r.startTestEditing(c)
}

func newLocalRunner() *localRunner {
	return &localRunner{client: api}
}

func (r *localRunner) startRun(c cliContext) error {
	var err error

	params, err := r.makeRunParams(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	defer r.client.DeleteEnvironment(params.EnvironmentID)

	runStatus, err := r.client.CreateRun(params)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	log.Printf("Run %v has been created.", runStatus.ID)

	return monitorRunStatus(c, runStatus.ID)
}

func (r *localRunner) startTestEditing(c cliContext) error {
	var err error

	params, err := r.makeRunParams(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	defer r.client.DeleteEnvironment(params.EnvironmentID)

	testId := params.Tests.([]int)[0]
	url := fmt.Sprintf("%v/tests/%v?envId=%v", rainforest.BaseURL, testId, params.EnvironmentID)
	openBrowserWithUrl(url)
	time.Sleep(20 * time.Second)

	return nil
}

func openBrowserWithUrl(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
		case "windows":
			// TODO: test on windows
			cmd = "cmd"
			args = []string{"/c", "start"}
		case "darwin":
			cmd = "open"
		default:
			// TODO: test on Linux
			cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func (r *localRunner) setupLocalEnvironment() (int, error) {
	rflocalHost := os.Getenv("RFLOCAL_HOST")
	if rflocalHost == "" {
		log.Print("RFLOCAL_HOST not set, falling back to localhost")
		rflocalHost = "localhost"
	}

	rflocalPortS := os.Getenv("RFLOCAL_PORT")
	if rflocalPortS == "" {
		log.Print("RFLOCAL_HOST not set, falling back to 3000")
		rflocalPortS = "3000"
	}
	rflocalPort, err := strconv.Atoi(rflocalPortS)
	if err != nil {
		log.Printf("Cannot use RFLOCAL_HOST value of %v", rflocalPortS)
		return -1, err
	}

	tunnelURL, err := r.openTunnel(rflocalHost, rflocalPort)
	if err != nil {
		return -1, err
	}

	environment, err := r.client.CreateTemporaryEnvironment(tunnelURL)
	if err != nil {
		return -1, err
	}

	return environment.ID, nil
}

func (r *localRunner) openTunnel(host string, port int) (string, error) {
	tunnel, err := GetTunnel(port, host)
	if err != nil {
		return "", err
	}
	log.Printf("Exposing %v:%v at %v", host, port, tunnel.URL())
	return tunnel.URL(), nil
}

func (r *localRunner) makeRunParams(c cliContext) (rainforest.RunParams, error) {
	// we can ignore rfml tests here, so just passing in an empty slice
	var localTests []*rainforest.RFTest
	// delegate most of the work to runner.makeRunParams
	params, err := newRunner().makeRunParams(c, localTests)
	if err != nil {
		return rainforest.RunParams{}, err
	}
	// override the environment with the tunnel
	environmentID, err := r.setupLocalEnvironment()
	if err != nil {
		return rainforest.RunParams{}, err
	}
	params.EnvironmentID = environmentID

	// also override the crowd, just in case
	params.Crowd = "automation"
	return params, nil
}
