// Copyright 2022 Anton Mamonov <hi@antonmamonov.com> GNU GENERAL PUBLIC LICENSE
package backup

import (
	"context"
	"flag"
	"fmt"
	"regexp"
	"time"

	"github.com/antonmamonov/k8s-backup-scheduler/k8sutils"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
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

	// create a new cluster role with the required permissions to access the source volume's pod in the namespace
	clusterRoleName := "backup-cluster-role"

	// check if the cluster role already exists
	_, getClusterRoleError := k8sConfig.ClientSet.RbacV1().ClusterRoles().Get(context.TODO(), clusterRoleName, metav1.GetOptions{})

	if getClusterRoleError != nil {
		// create the cluster role
		_, createClusterRoleError := k8sConfig.ClientSet.RbacV1().ClusterRoles().Create(context.TODO(), &rbac.ClusterRole{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRole",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterRoleName,
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"pods/exec"},
					Verbs:     []string{"create"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"pods", "pods/log"},
					Verbs:     []string{"get", "list"},
				},
			},
		}, metav1.CreateOptions{})

		if createClusterRoleError != nil {
			return createClusterRoleError
		}

		fmt.Println("[BackupVolume] Created new ClusterRole:", clusterRoleName)
	}

	// create the service account
	serviceAccountName := "backup-service-account"

	// check if the service account already exists
	_, getServiceAccountError := k8sConfig.ClientSet.CoreV1().ServiceAccounts(k8sVolumeBackupConfig.DestinationVolumeNamespace).Get(context.TODO(), serviceAccountName, metav1.GetOptions{})

	if getServiceAccountError != nil {
		// create the service account
		_, createServiceAccountError := k8sConfig.ClientSet.CoreV1().ServiceAccounts(k8sVolumeBackupConfig.DestinationVolumeNamespace).Create(context.TODO(), &v1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceAccountName,
			},
		}, metav1.CreateOptions{})

		if createServiceAccountError != nil {
			return createServiceAccountError
		}

		fmt.Println("[BackupVolume] Created new ServiceAccount:", serviceAccountName, "in namespace:", k8sVolumeBackupConfig.DestinationVolumeNamespace)
	}

	// create the cluster role binding
	clusterRoleBindingName := "backup-cluster-role-binding"

	// check if the cluster role binding already exists
	_, getClusterRoleBindingError := k8sConfig.ClientSet.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})

	if getClusterRoleBindingError != nil {
		// create the cluster role binding
		_, createClusterRoleBindingError := k8sConfig.ClientSet.RbacV1().ClusterRoleBindings().Create(context.TODO(), &rbac.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterRoleBindingName,
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      serviceAccountName,
					Namespace: k8sVolumeBackupConfig.DestinationVolumeNamespace,
				},
			},
			RoleRef: rbac.RoleRef{
				Kind:     "ClusterRole",
				Name:     clusterRoleName,
				APIGroup: "rbac.authorization.k8s.io",
			},
		}, metav1.CreateOptions{})

		if createClusterRoleBindingError != nil {
			return createClusterRoleBindingError
		}

		fmt.Println("[BackupVolume] Created new ClusterRoleBinding:", clusterRoleBindingName)
	}

	time.Sleep(1 * time.Second)

	// find the latest backup service token secret
	backupServiceRegex := regexp.MustCompile(serviceAccountName + `-token-.*`)
	backupServiceTokenSecretName := ""

	// get the list of secrets in the namespace
	secrets, listSecretsError := k8sConfig.ClientSet.CoreV1().Secrets(k8sVolumeBackupConfig.DestinationVolumeNamespace).List(context.TODO(), metav1.ListOptions{})
	if listSecretsError != nil {
		return listSecretsError
	}

	// loop through the secrets and find the latest backup service token secret
	for _, secret := range secrets.Items {
		if backupServiceRegex.MatchString(secret.Name) {
			backupServiceTokenSecretName = secret.Name
		}
	}

	// check if the service account token secret already exists

	// create a new job with the destination volume attached

	jobName := "backup-job-" + k8sVolumeBackupConfig.DestinationVolumeName

	jobVolumeMounts := []v1.VolumeMount{
		{
			Name:      "destination-volume",
			MountPath: "/backup",
		},
		{
			Name:      "backup-service-account-token",
			MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
		},
	}

	jobVolumes := []v1.Volume{
		{
			Name: "destination-volume",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: k8sVolumeBackupConfig.DestinationVolumeName,
				},
			},
		},
		{
			Name: "backup-service-account-token",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: backupServiceTokenSecretName,
				},
			},
		},
	}

	SourcePodName := ""
	SourcePodNamespace := k8sVolumeBackupConfig.SourceVolumeNamespace
	SourcePodDirectory := ""
	SourcePodVolumeMountName := ""
	DestinationDirectory := "/backup"

	// find pod that has the volume claim attached
	pods, listPodsError := k8sConfig.ClientSet.CoreV1().Pods(k8sVolumeBackupConfig.SourceVolumeNamespace).List(context.TODO(), metav1.ListOptions{})

	if listPodsError != nil {
		return listPodsError
	}

	// loop through the pods and find the pod that has the volume claim attached
	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == k8sVolumeBackupConfig.SourceVolumeName {
				SourcePodName = pod.Name
				SourcePodNamespace = pod.Namespace
				SourcePodVolumeMountName = volume.Name
			}
		}
	}

	// through the pods and find the volume mount that has the volume claim attached
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == SourcePodVolumeMountName {
					SourcePodDirectory = volumeMount.MountPath
				}
			}
		}
	}

	// pass as environment variables to the backup job

	syncK8sEnvVar := []v1.EnvVar{
		{
			Name:  "SOURCE_POD_NAME",
			Value: SourcePodName,
		},
		{
			Name:  "SOURCE_POD_NAMESPACE",
			Value: SourcePodNamespace,
		},
		{
			Name:  "SOURCE_POD_DIRECTORY",
			Value: SourcePodDirectory,
		},
		{
			Name:  "DESTINATION_DIRECTORY",
			Value: DestinationDirectory,
		},
	}

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
			BackoffLimit: func() *int32 { i := int32(7); return &i }(),
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: "OnFailure",
					Containers: []v1.Container{
						{
							Name:  "main",
							Image: "antonm/k8s-pv-backup-scheduler",

							Args: []string{
								"sync",
							},

							// Quick Development
							// Command: []string{"/bin/bash"},
							// Args: []string{
							// 	"-c",
							// 	"sleep 9999;",
							// },
							VolumeMounts: jobVolumeMounts,
							Env:          syncK8sEnvVar,
						},
					},
					Volumes: jobVolumes,
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
