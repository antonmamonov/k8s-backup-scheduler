// Copyright 2022 Anton Mamonov <hi@antonmamonov.com> GNU GENERAL PUBLIC LICENSE
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/antonmamonov/k8s-backup-scheduler/k8sutils"
	"github.com/cristalhq/acmd"
)

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
