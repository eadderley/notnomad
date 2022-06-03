package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	nomadapi "github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/kyokomi/emoji/v2"
)

var wg sync.WaitGroup

func main() {

	// nomadclient automatically accepts env vars
	_, nAddrExists := os.LookupEnv("NOMAD_ADDR")
	_, nTokenExists := os.LookupEnv("NOMAD_TOKEN")

	if !nAddrExists {
		fmt.Println("NOMAD_ADDR must be set!")
		os.Exit(1)
	}

	if !nTokenExists {
		fmt.Println("NOMAD_TOKEN must be set!")
		os.Exit(1)
	}

	nomadClient, err := nomadapi.NewClient(nomadapi.DefaultConfig())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hclfilePath := os.Args[1]
	nomadJob, err := jobspec.ParseFile(hclfilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// api equivalent of nomad job run
	response, _, err := nomadClient.Jobs().Register(nomadJob, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	emoji.Println(fmt.Sprintf(":check_mark: Job at %s run successfully - EvalID:%s", hclfilePath, response.EvalID))

	var query nomadapi.QueryOptions
	var deploy []*nomadapi.Deployment

	for ok := true; ok; ok = len(deploy) == 0 {
		deploy, _, err = nomadClient.Jobs().Deployments(*nomadJob.Name, false, &query)
	}

	alloc, _, err := nomadClient.Jobs().Allocations(*nomadJob.ID, false, &query)
	for _, id := range alloc {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allocCommand := exec.Command("nomad", "alloc", "logs", id.ID)
			stdout, err := allocCommand.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(string(stdout))
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		depCommand := exec.Command("nomad", "deployment", "status", "-monitor", deploy[0].ID)
		depReader, err := depCommand.StdoutPipe()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		scanner := bufio.NewScanner(depReader)
		go func() {
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}()
		if err := depCommand.Start(); err != nil {
			log.Fatal(err)
		}
		if err := depCommand.Wait(); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Printf("allocID: %v\n", alloc[0].ID)

	wg.Wait()
}
