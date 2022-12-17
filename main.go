// Copyright 2022 Anton Mamonov <hi@antonmamonov.com> GNU GENERAL PUBLIC LICENSE
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/antonmamonov/k8s-backup-scheduler/backup"
	"github.com/antonmamonov/k8s-backup-scheduler/k8sutils"
	"github.com/antonmamonov/k8s-backup-scheduler/sync"
	"github.com/cristalhq/acmd"
)

type generalFlags struct {
	IsVerbose                  bool
	SourceVolumeName           string
	SourceVolumeNamespace      string
	DestinationVolumeName      string
	DestinationVolumeNamespace string
}

func (c *generalFlags) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&c.IsVerbose, "verbose", false, "should be verbose")
	fs.StringVar(&c.SourceVolumeName, "sourcevolumename", "pvc-1", "The source persistent volume name")
	fs.StringVar(&c.SourceVolumeNamespace, "sourcevolumenamespace", "default", "The source persistent volume namespace")
	fs.StringVar(&c.DestinationVolumeName, "destinationvolumename", "pvc-2", "The destination persistent volume name")
	fs.StringVar(&c.DestinationVolumeNamespace, "destinationvolumenamespace", "default", "The destination persistent volume namespace")
	return fs
}

type commandFlags struct {
	generalFlags
	File string
}

func (c *commandFlags) Flags() *flag.FlagSet {
	fs := c.generalFlags.Flags()
	fs.StringVar(&c.File, "file", "input.txt", "file to process")
	return fs
}

func main() {

	cmds := []acmd.Command{
		{
			Name:        "now",
			Description: "prints current time",
			ExecFunc: func(ctx context.Context, args []string) error {
				fmt.Printf("now: %s\n", time.Now())
				return nil
			},
		},
		{
			Name:        "k8s-test",
			Description: "Test K8s Cluster Config Connection",
			ExecFunc: func(ctx context.Context, args []string) error {
				isGood, isError := k8sutils.CheckK8sClusterConfigConnection()

				if isError != nil {
					return isError
				}

				if isGood {
					fmt.Printf("K8s Cluster Config Connection is GOOD! You're ready to use the backup!\n")
				} else {
					fmt.Printf("K8s Cluster Config Connection is BAD! Please check config at /root/.kube/config\n")
				}
				return nil
			},
		},
		{
			Name:        "backup",
			Description: "Back up a persistent volume to a source volume",
			ExecFunc: func(ctx context.Context, args []string) error {

				// get command flags
				var cfg backup.BackupVolumeFlags
				if err := cfg.Flags().Parse(args); err != nil {
					return err
				}

				backupVolumeError := backup.BackupVolume(&cfg)

				if backupVolumeError != nil {
					return backupVolumeError
				}

				return nil
			},
		},
		{
			Name:        "sync",
			Description: "synchronize a persistent volume to a destination folder",
			ExecFunc: func(ctx context.Context, args []string) error {

				// get command flags
				var cfg sync.SyncVolumeFlags
				if err := cfg.Flags().Parse(args); err != nil {
					return err
				}

				if os.Getenv("SOURCE_POD_NAME") != "" {
					cfg.SourcePodName = os.Getenv("SOURCE_POD_NAME")
				}

				if os.Getenv("SOURCE_POD_NAMESPACE") != "" {
					cfg.SourcePodNamespace = os.Getenv("SOURCE_POD_NAMESPACE")
				}

				if os.Getenv("SOURCE_POD_DIRECTORY") != "" {
					cfg.SourcePodDirectory = os.Getenv("SOURCE_POD_DIRECTORY")
				}

				if os.Getenv("DESTINATION_DIRECTORY") != "" {
					cfg.DestinationDirectory = os.Getenv("DESTINATION_DIRECTORY")
				}

				syncVolumeError := sync.SyncVolume(&cfg)

				if syncVolumeError != nil {
					return syncVolumeError
				}

				return nil
			},
		},
	}

	// all the acmd.Config fields are optional
	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName:        "k8s-backup-scheduler",
		AppDescription: "Simple Kubernetes Persistent Volume Backup with Scheduling <hi@antonmamonov.com>",
		Version:        "v0.0.1",
	})

	if err := r.Run(); err != nil {
		r.Exit(err)
	}
}
