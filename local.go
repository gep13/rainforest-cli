package main

import (
	"fmt"
	"log"

	"github.com/rainforestapp/rainforest-cli/rainforest"
	"github.com/urfave/cli"

	"github.com/localtunnel/go-localtunnel"
)

func GetTunnel(port int, host string) (*localtunnel.LocalTunnel, error) {

	tunnel, err := localtunnel.New(port, host, localtunnel.Options{})

	if err != nil {
		fmt.Println("Error: %v", err)
		return tunnel, err
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

func newLocalRunner() *localRunner {
	return &localRunner{client: api}
}

func (r *localRunner) startTunnel(c cliContext) error {
	tunnel, _ := GetTunnel(8000, "localhost")
	customURL := tunnel.URL()

	environment, err := r.client.CreateTemporaryEnvironment(customURL)
	if err != nil {
		return err
	}

}

func (r *localRunner) startRun(c cliContext) error {
	var err error

	params, err := r.makeRunParams(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	runStatus, err := r.client.CreateRun(params)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	log.Printf("Run %v has been created.", runStatus.ID)

	return monitorRunStatus(c, runStatus.ID)
}

func (r *localRunner) makeRunParams(c cliContext) (rainforest.RunParams, error) {
	crowd := "automation"

	browsers := []string{"chrome_1440_900"}
	expandedBrowsers := expandStringSlice(browsers)

	// open localtunnel
	tunnel, _ := GetTunnel(8000, "localhost")
	customURL := tunnel.URL()

	environment, err := r.client.CreateTemporaryEnvironment(customURL)
	if err != nil {
		return rainforest.RunParams{}, err
	}

	var testIDs interface{}
	testIDs = []int{
		273861,
	}

	return rainforest.RunParams{
		Tests:         testIDs,
		Crowd:         crowd,
		Browsers:      expandedBrowsers,
		EnvironmentID: environment.ID,
	}, nil
}
