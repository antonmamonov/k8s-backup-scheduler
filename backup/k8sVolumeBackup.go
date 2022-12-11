// Copyright 2022 Anton Mamonov <hi@antonmamonov.com> GNU GENERAL PUBLIC LICENSE
package backup

import (
	"context"
	"fmt"

	"github.com/antonmamonov/k8s-backup-scheduler/k8sutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sVolumeBackupConfig struct {
	SourceVolumeName           string
	SourceVolumeNamespace      string
	DestinationVolumeName      string
	DestinationVolumeNamespace string
}

func BackupVolume(k8sVolumeBackupConfig *K8sVolumeBackupConfig) error {

	k8sConfig := &k8sutils.K8sClusterConfig{}

	// get status of persistent source volume using the k8sConfig
	pVolume, getPVolumeError := k8sConfig.ClientSet.CoreV1().PersistentVolumes().Get(context.TODO(), k8sVolumeBackupConfig.SourceVolumeName, metav1.GetOptions{})

	if getPVolumeError != nil {
		return getPVolumeError
	}

	// get status of pVolumes
	fmt.Printf("pVolume: %s\n", &pVolume.Status)

	return nil
}
