package dataset

import (
	"context"
	"os"
	"testing"

	"github.com/stratastor/logger"
	"github.com/stratastor/rodent/pkg/errors"
	"github.com/stratastor/rodent/pkg/zfs/command"
	"github.com/stratastor/rodent/pkg/zfs/pool"
	"github.com/stratastor/rodent/pkg/zfs/testutil"
)

func TestDataTransferOperations(t *testing.T) {

	logLevel := os.Getenv("RODENT_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	// Setup test environment
	env := testutil.NewTestEnv(t, 3)
	defer env.Cleanup()

	executor := command.NewCommandExecutor(true, logger.Config{LogLevel: logLevel})
	poolMgr := pool.NewManager(executor)
	datasetMgr := NewManager(executor)

	// Create source pool
	srcPoolName := testutil.GeneratePoolName()
	srcPoolCfg := pool.CreateConfig{
		Name: srcPoolName,
		VDevSpec: []pool.VDevSpec{
			{
				Type:    "raidz",
				Devices: env.GetLoopDevices(),
			},
		},
	}

	err := poolMgr.Create(context.Background(), srcPoolCfg)
	if err != nil {
		t.Fatalf("failed to create source pool: %v", err)
	}

	// Track pool state for cleanup
	var srcPoolDestroyed bool
	defer func() {
		if !srcPoolDestroyed {
			poolMgr.Destroy(context.Background(), srcPoolName, true)
		}
	}()

	// Create destination pool
	env2 := testutil.NewTestEnv(t, 3)
	defer env2.Cleanup()

	dstPoolName := testutil.GeneratePoolName()
	dstPoolCfg := pool.CreateConfig{
		Name: dstPoolName,
		VDevSpec: []pool.VDevSpec{
			{
				Type:    "raidz",
				Devices: env2.GetLoopDevices(),
			},
		},
	}

	err = poolMgr.Create(context.Background(), dstPoolCfg)
	if err != nil {
		t.Fatalf("failed to create destination pool: %v", err)
	}

	var dstPoolDestroyed bool
	defer func() {
		if !dstPoolDestroyed {
			poolMgr.Destroy(context.Background(), dstPoolName, true)
		}
	}()

	t.Run("BasicTransfer", func(t *testing.T) {
		// Create source filesystem with data
		srcFs := "/fs1"
		srcDs := srcPoolName + srcFs
		err := datasetMgr.Create(context.Background(), CreateConfig{
			Name: srcDs,
			Type: "filesystem",
			Properties: map[string]string{
				"compression": "on",
			},
		})
		if err != nil {
			t.Fatalf("failed to create source filesystem: %v", err)
		}

		// Create test data in source
		// TODO: Add test data creation

		// Create snapshot
		snapStr := srcFs + "@snap1"
		snapName := srcPoolName + snapStr
		err = datasetMgr.CreateSnapshot(context.Background(), SnapshotConfig{
			Dataset: srcDs,
			Name:    "snap1",
		})
		if err != nil {
			t.Fatalf("failed to create snapshot: %v", err)
		}

		// Test local transfer
		dstFs := dstPoolName + "/received"
		err = datasetMgr.Create(context.Background(), CreateConfig{
			Name: dstFs,
			Type: "filesystem",
			Properties: map[string]string{
				"compression": "on",
			},
		})
		if err != nil {
			t.Fatalf("failed to create source filesystem: %v", err)
		}

		dstDs := dstFs + snapStr

		err = datasetMgr.SendReceive(context.Background(),
			SendConfig{
				Snapshot:   snapName,
				Compressed: true,
				LogLevel:   logLevel,
			},
			ReceiveConfig{
				Target:     dstFs,
				Force:      true,
				Properties: map[string]string{"compression": "on"},
				Resumable:  false,
				UseParent:  true,
			},
		)
		if err != nil {
			if rerr, ok := err.(*errors.RodentError); ok {
				t.Fatalf("failed remote transfer: %v\nOutput: %s\nCommand: %s",
					rerr,
					rerr.Metadata["output"],
					rerr.Metadata["command"])
			}
			t.Fatalf("failed remote transfer: %v", err)
		}

		// Verify destination exists
		exists, err := datasetMgr.Exists(context.Background(), dstDs)
		if err != nil {
			t.Fatalf("failed to check destination: %v", err)
		}
		if !exists {
			t.Error("destination dataset does not exist")
		}

		// Verify properties were transferred
		prop, err := datasetMgr.GetProperty(context.Background(), dstFs+srcFs, "compression")
		if err != nil {
			t.Fatalf("failed to get property: %v", err)
		}
		if prop.Value != "on" {
			t.Errorf("property value = %v, want 'on'", prop.Value)
		}
	})

	t.Run("IncrementalTransfer", func(t *testing.T) {
		srcFs := srcPoolName + "/fs2"
		err := datasetMgr.Create(context.Background(), CreateConfig{
			Name: srcFs,
			Type: "filesystem",
		})
		if err != nil {
			t.Fatalf("failed to create source filesystem: %v", err)
		}

		// Create initial snapshot
		snap1 := srcFs + "@snap1"
		err = datasetMgr.CreateSnapshot(context.Background(), SnapshotConfig{
			Dataset: srcFs,
			Name:    "snap1",
		})
		if err != nil {
			t.Fatalf("failed to create first snapshot: %v", err)
		}

		// Initial transfer
		dstFs := dstPoolName + "/fs2"
		err = datasetMgr.SendReceive(context.Background(),
			SendConfig{
				Snapshot:   snap1,
				Properties: true,
				LogLevel:   logLevel,
			},
			ReceiveConfig{
				Target:     dstFs,
				Force:      true,
				Properties: map[string]string{"compression": "on"},
			},
		)
		if err != nil {
			t.Fatalf("failed initial transfer: %v", err)
		}

		// Create second snapshot
		snap2 := srcFs + "@snap2"
		err = datasetMgr.CreateSnapshot(context.Background(), SnapshotConfig{
			Dataset: srcFs,
			Name:    "snap2",
		})
		if err != nil {
			t.Fatalf("failed to create second snapshot: %v", err)
		}

		// Incremental transfer
		err = datasetMgr.SendReceive(context.Background(),
			SendConfig{
				Snapshot:     snap2,
				FromSnapshot: snap1,
				Properties:   true,
			},
			ReceiveConfig{
				Target:     dstFs,
				Force:      true,
				Properties: map[string]string{"compression": "on"},
			},
		)
		if err != nil {
			t.Fatalf("failed incremental transfer: %v", err)
		}

		// Verify both snapshots exist on destination
		for _, snap := range []string{dstFs + "@snap1", dstFs + "@snap2"} {
			exists, err := datasetMgr.Exists(context.Background(), snap)
			if err != nil {
				t.Fatalf("failed to check snapshot %s: %v", snap, err)
			}
			if !exists {
				t.Errorf("snapshot %s does not exist", snap)
			}
		}
	})

	t.Run("ResumeTransfer", func(t *testing.T) {
		srcFs := srcPoolName + "/fs3"
		err := datasetMgr.Create(context.Background(), CreateConfig{
			Name: srcFs,
			Type: "filesystem",
			Properties: map[string]string{
				"compression": "on",
			},
		})
		if err != nil {
			t.Fatalf("failed to create source filesystem: %v", err)
		}

		// Create large test data
		// TODO: Add large test data creation

		// Create snapshot
		snapName := srcFs + "@snap1"
		err = datasetMgr.CreateSnapshot(context.Background(), SnapshotConfig{
			Dataset: srcFs,
			Name:    "snap1",
		})
		if err != nil {
			t.Fatalf("failed to create snapshot: %v", err)
		}

		// Start resumable receive
		dstFs := dstPoolName + "/fs3"
		err = datasetMgr.SendReceive(context.Background(),
			SendConfig{
				Snapshot:   snapName,
				Properties: true,
				Progress:   true,
				LogLevel:   logLevel,
			},
			ReceiveConfig{
				Target:    dstFs,
				Force:     true,
				Resumable: true, // Enable resumable receive
			},
		)
		// TODO: Cancel transfer immediately to simulate failure
		if err != nil {
			// Get resume token
			token, err := datasetMgr.GetResumeToken(context.Background(), dstFs)
			if err != nil {
				t.Fatalf("failed to get resume token: %v", err)
			}
			if token == "" {
				t.Fatal("no resume token available")
			}

			// Resume transfer
			err = datasetMgr.SendReceive(context.Background(),
				SendConfig{
					ResumeToken: token,
					Progress:    true,
					Verbose:     true,
					LogLevel:    logLevel,
				},
				ReceiveConfig{
					Target:    dstFs,
					Force:     true,
					Resumable: true,
				},
			)
			if err != nil {
				t.Fatalf("failed to resume transfer: %v", err)
			}
		}

		// Verify transfer completed
		exists, err := datasetMgr.Exists(context.Background(), dstFs)
		if err != nil {
			t.Fatalf("failed to check destination: %v", err)
		}
		if !exists {
			t.Error("destination dataset does not exist")
		}

		// Verify properties
		prop, err := datasetMgr.GetProperty(context.Background(), dstFs, "compression")
		if err != nil {
			t.Fatalf("failed to get property: %v", err)
		}
		if prop.Value != "on" {
			t.Errorf("property value = %v, want 'on'", prop.Value)
		}
	})

	t.Run("RemoteTransfer", func(t *testing.T) {
		// sudo RODENT_REMOTE_TEST="yes" RODENT_REMOTE_DATASET="tank/ds3" RODENT_REMOTE_HOST="13.xx.xx.xxx" RODENT_REMOTE_USER="ubuntu" RODENT_REMOTE_PRIVATE_KEY="/home/ubuntu/.ssh/id_ed25519" go test -v -run TestDataTransferOperations/RemoteTransfer

		// Skip if no remote test environment
		if os.Getenv("RODENT_REMOTE_TEST") == "" {
			t.Skip("Skipping remote transfer test")
		}

		// Create source filesystem with data
		srcFs := srcPoolName + "/fs4"
		err := datasetMgr.Create(context.Background(), CreateConfig{
			Name: srcFs,
			Type: "filesystem",
			Properties: map[string]string{
				"compression": "on",
			},
		})
		if err != nil {
			t.Fatalf("failed to create source filesystem: %v", err)
		}

		// Create test data in source

		// Create snapshot
		snapName := srcFs + "@snap1"
		err = datasetMgr.CreateSnapshot(context.Background(), SnapshotConfig{
			Dataset: srcFs,
			Name:    "snap1",
		})
		if err != nil {
			t.Fatalf("failed to create snapshot: %v", err)
		}

		// Test remote replication
		remoteConfig := RemoteConfig{
			Host:             os.Getenv("RODENT_REMOTE_HOST"),
			User:             os.Getenv("RODENT_REMOTE_USER"),
			Port:             22,
			PrivateKey:       os.Getenv("RODENT_REMOTE_PRIVATE_KEY"),
			SkipHostKeyCheck: true,
		}

		t.Logf("Remote config: host=%s, user=%s, key=%s",
			remoteConfig.Host,
			remoteConfig.User,
			remoteConfig.PrivateKey)

		err = datasetMgr.SendReceive(context.Background(),
			SendConfig{
				Snapshot:   snapName,
				Compressed: true,
				LogLevel:   logLevel,
				Progress:   true,
			},
			ReceiveConfig{
				Target:       os.Getenv("RODENT_REMOTE_DATASET"),
				Force:        true,
				Resumable:    true,
				UseParent:    true,
				RemoteConfig: remoteConfig,
			},
		)
		if err != nil {
			if rerr, ok := err.(*errors.RodentError); ok {
				t.Fatalf("failed remote transfer: %v\nOutput: %s\nCommand: %s",
					rerr,
					rerr.Metadata["output"],
					rerr.Metadata["command"])
			}
			t.Fatalf("failed remote transfer: %v", err)
		}

	})

	// Clean up
	for _, name := range []string{srcPoolName, dstPoolName} {
		if err := poolMgr.Destroy(context.Background(), name, true); err != nil {
			t.Errorf("failed to destroy pool %s: %v", name, err)
		}
	}
	srcPoolDestroyed = true
	dstPoolDestroyed = true
}
