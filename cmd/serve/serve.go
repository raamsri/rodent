package serve

import (
	"context"
	"log"

	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
	"github.com/stratastor/rodent/config"
	"github.com/stratastor/rodent/internal/constants"
	"github.com/stratastor/rodent/pkg/lifecycle"
	"github.com/stratastor/rodent/pkg/server"
)

var detached bool

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Rodent server",
		Run:   runServe,
	}

	cmd.Flags().BoolVarP(&detached, "detach", "d", false, "Run as a daemon")
	return cmd
}

func runServe(cmd *cobra.Command, args []string) {
	rc := config.GetConfig()
	pidFile := constants.RodentPIDFilePath
	// Check for existing instance before proceeding
	if err := lifecycle.EnsureSingleInstance(pidFile); err != nil {
		log.Fatalf("Failed to start: %v\n", err)
	}

	if detached {
		ctx := &daemon.Context{
			PidFileName: pidFile,
			PidFilePerm: 0644,
			LogFileName: rc.Logs.Path,
			LogFilePerm: 0640,
			WorkDir:     "/",
			Umask:       027,
			Args:        []string{"rodent", "serve"},
		}

		d, err := ctx.Reborn()
		if err != nil {
			log.Fatalf("Failed to start daemon: %v\n", err)
		}

		if d != nil {
			log.Println("Rodent is running as a daemon")
			return
		}
		defer ctx.Release()
	}

	startServer()
}

func startServer() {
	cfg := config.GetConfig()

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register the context canceller
	lifecycle.RegisterContextCanceller(cancel)

	// Register shutdown hook for server cleanup
	lifecycle.RegisterShutdownHook(func() {
		log.Println("Shutting down server")
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v\n", err)
		}
	})

	// Start handling lifecycle signals (e.g., SIGTERM, SIGHUP)
	go lifecycle.HandleSignals(ctx)

	// Start the server
	log.Printf("Starting Rodent server on port %d\n", cfg.Server.Port)
	if err := server.Start(ctx, cfg.Server.Port); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}
