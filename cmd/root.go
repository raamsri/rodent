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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stratastor/rodent/cmd/config"
	"github.com/stratastor/rodent/cmd/health"
	"github.com/stratastor/rodent/cmd/logs"
	"github.com/stratastor/rodent/cmd/serve"
	"github.com/stratastor/rodent/cmd/status"
	"github.com/stratastor/rodent/cmd/version"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "rodent",
		Short: "Rodent: StrataSTOR Node Agent",
	}

	rootCmd.AddCommand(serve.NewServeCmd())
	rootCmd.AddCommand(version.NewVersionCmd())
	rootCmd.AddCommand(health.NewHealthCmd())
	rootCmd.AddCommand(status.NewStatusCmd())
	rootCmd.AddCommand(logs.NewLogsCmd())
	rootCmd.AddCommand(config.NewConfigCmd())

	return rootCmd
}
