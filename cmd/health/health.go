package health

import (
	"github.com/spf13/cobra"
	"github.com/stratastor/rodent/config"
	"github.com/stratastor/rodent/pkg/health"
)

func NewHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check Rodent health",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.GetConfig() // cfg shoudln't be nil
			checker := health.NewHealthChecker(cfg)
			return checker.CheckHealth()
		},
	}
}