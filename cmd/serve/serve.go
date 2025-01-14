/*
 * Copyright 2024-2025 Raamsri Kumar <raam@tinkershack.in>
 * Copyright 2024-2025 The StrataSTOR Authors and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package serve

import (
	"context"
	"fmt"
	"os"

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
		fmt.Printf("Failed to start: %v\n", err)
		os.Exit(1)
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
			fmt.Printf("Failed to start daemon: %v\n", err)
			os.Exit(1)
		}

		if d != nil {
			fmt.Println("Rodent is running as a daemon")
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
		fmt.Println("Shutting down server")
		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Error during server shutdown: %v\n", err)
		}
	})

	// Start handling lifecycle signals (e.g., SIGTERM, SIGHUP)
	go lifecycle.HandleSignals(ctx)

	// Start the server
	fmt.Printf("Starting Rodent server on port %d\n", cfg.Server.Port)
	if err := server.Start(ctx, cfg.Server.Port); err != nil {
		fmt.Printf("Failed to start server: %v", err)
	}
}
