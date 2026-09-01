package cmd

import (
	"ai-monitoring/config"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

var bootCmd = &wcli.Command{
	Use:   "boot",
	Short: "Analyze recent boot logs",
	Long:  "Fetch system boot logs and diagnose them with the LLM.",
	Run: func(ctx *wcli.Context) error {
		analysis, err := runBootDiagnosis(ctx, config.GlobalConfig)
		if err != nil {
			return err
		}
		if analysis == "" {
			rich.Println("[green]탐지된 부팅 에러나 특이사항이 없습니다.[/green]")
			return nil
		}
		rich.Println("[bold][white]Boot Diagnosis:[/white][/bold]")
		rich.Println("[white]%s[/white]", analysis)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bootCmd)
}
