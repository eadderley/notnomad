package cmd

import (
	"fmt"
	"os"
	"os/exec"

	nomadapi "github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/kyokomi/emoji/v2"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs job(s) on nomad",
	Long: `Runs job(s) on a nomad cluster given hcl file(s).
	
Examples:
notnomad run --file echo.hcl 
notnomad run -f echo.hcl
Run the echo.hcl 
`,
	Run: func(cmd *cobra.Command, args []string) {

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

		hclfilePath := cmd.Flag("file").Value.String()
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

		depCommand := exec.Command("nomad", "deployment", "status", deploy[0].ID)
		stdout, err := depCommand.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(string(stdout))

		alloc, _, err := nomadClient.Jobs().Allocations(*nomadJob.ID, false, &query)
		fmt.Printf("allocID: %v\n", alloc[0].ID)

		allocCommand := exec.Command("nomad", "alloc", "logs", alloc[0].ID)
		stdout, err = allocCommand.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(string(stdout))

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("file", "f", "", "Run a supported job hcl")
	runCmd.MarkFlagRequired("file")
}
