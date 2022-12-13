// Copyright 2022 Anton Mamonov <hi@antonmamonov.com> GNU GENERAL PUBLIC LICENSE
package backup

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/antonmamonov/k8s-backup-scheduler/k8sutils"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackupVolumeFlags struct {
	DoOverride                 bool
	SourceVolumeName           string
	SourceVolumeNamespace      string
	DestinationVolumeName      string
	DestinationVolumeNamespace string
}

func (c *BackupVolumeFlags) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&c.SourceVolumeName, "sourcevolumename", "pvc-1", "The source persistent volume claim name")
	fs.StringVar(&c.SourceVolumeNamespace, "sourcevolumenamespace", "default", "The source persistent volume claim namespace")
	fs.StringVar(&c.DestinationVolumeName, "destinationvolumename", "pvc-2", "The destination persistent volume claim name")
	fs.StringVar(&c.DestinationVolumeNamespace, "destinationvolumenamespace", "default", "The destination persistent volume claim namespace")
	fs.BoolVar(&c.DoOverride, "override", false, "Override and delete existing destination volume (WARNING: THIS WILL DELETE ALL DATA IN DESTINATION VOLUME)")
	return fs
}

func BackupVolume(k8sVolumeBackupConfig *BackupVolumeFlags) error {

	k8sConfig, getK8sClusterConfigError := k8sutils.GetK8sClusterConfig()

	if getK8sClusterConfigError != nil {
		return getK8sClusterConfigError
	}

	// get status of persistent source volume using the k8sConfig
	sourcePVC, getPVolumeError := k8sConfig.ClientSet.CoreV1().PersistentVolumeClaims(k8sVolumeBackupConfig.SourceVolumeNamespace).Get(context.TODO(), k8sVolumeBackupConfig.SourceVolumeName, metav1.GetOptions{})

	if getPVolumeError != nil {
		return getPVolumeError
	}

	// check if it's bound
	if sourcePVC.Status.Phase != "Bound" {
		return fmt.Errorf("pVolume %s is not bound", k8sVolumeBackupConfig.SourceVolumeName)
	}

	newPVCClaim := v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sVolumeBackupConfig.DestinationVolumeName,
			Namespace: k8sVolumeBackupConfig.DestinationVolumeNamespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: sourcePVC.Spec.AccessModes,
			Resources: v1.ResourceRequirements{
				Requests: sourcePVC.Spec.Resources.Requests,
			},
			StorageClassName: sourcePVC.Spec.StorageClassName,
		},
	}

	// if override is set to true, delete the destination volume if it exists
	if k8sVolumeBackupConfig.DoOverride {
		_, getDestinationPVolumeError := k8sConfig.ClientSet.CoreV1().PersistentVolumeClaims(k8sVolumeBackupConfig.DestinationVolumeNamespace).Get(context.TODO(), k8sVolumeBackupConfig.DestinationVolumeName, metav1.GetOptions{})
		if getDestinationPVolumeError == nil {
			// delete the destination volume
			deleteDestinationPVolumeError := k8sConfig.ClientSet.CoreV1().PersistentVolumeClaims(k8sVolumeBackupConfig.DestinationVolumeNamespace).Delete(context.TODO(), k8sVolumeBackupConfig.DestinationVolumeName, metav1.DeleteOptions{})
			if deleteDestinationPVolumeError != nil {
				return deleteDestinationPVolumeError
			}
		}
	}

	// create the destination persistent volume claim using the k8sConfig
	_, createPVolumeError := k8sConfig.ClientSet.CoreV1().PersistentVolumeClaims(k8sVolumeBackupConfig.DestinationVolumeNamespace).Create(context.TODO(), &newPVCClaim, metav1.CreateOptions{})

	if createPVolumeError != nil {
		return createPVolumeError
	}

	fmt.Println("[BackupVolume] Created new PVC:", k8sVolumeBackupConfig.DestinationVolumeName, "in namespace:", k8sVolumeBackupConfig.DestinationVolumeNamespace)

	// sleep for a few seconds to allow the volume to be created
	time.Sleep(3 * time.Second)
	// create a new job with the destination volume attached

	jobName := "backup-job-" + k8sVolumeBackupConfig.DestinationVolumeName

	syncK8sJob := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: k8sVolumeBackupConfig.DestinationVolumeNamespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: "OnFailure",
					Containers: []v1.Container{
						{
							Name:    "main",
							Image:   "antonm/k8s-pv-backup-scheduler",
							Command: []string{"/bin/bash"},
							Args: []string{
								"-c",
								"sleep 9999;",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "destination-volume",
									MountPath: "/backup",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "destination-volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: k8sVolumeBackupConfig.DestinationVolumeName,
								},
							},
						},
					},
				},
			},
		},
	}

	// create the syncK8sJob
	_, createSyncK8sJobError := k8sConfig.ClientSet.BatchV1().Jobs(k8sVolumeBackupConfig.DestinationVolumeNamespace).Create(context.TODO(), &syncK8sJob, metav1.CreateOptions{})

	if createSyncK8sJobError != nil {
		return createSyncK8sJobError
	}

	fmt.Println("[BackupVolume] Created new Job:", jobName, "in namespace:", k8sVolumeBackupConfig.DestinationVolumeNamespace)

	return nil
}
