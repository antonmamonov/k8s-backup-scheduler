package sync

import (
	"flag"
	"fmt"

	"github.com/antonmamonov/k8s-backup-scheduler/k8sutils"
)

type SyncVolumeFlags struct {
	SourcePodName        string
	SourcePodNamespace   string
	SourcePodDirectory   string
	DestinationDirectory string
}

func (c *SyncVolumeFlags) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&c.SourcePodName, "sourcepodname", "pod-1", "The source pod name to sync the volume from")
	fs.StringVar(&c.SourcePodNamespace, "sourcepodnamespace", "default", "The source pod namespace")
	fs.StringVar(&c.SourcePodDirectory, "sourcepoddirectory", "/data", "The source directory to sync the volume data from")
	fs.StringVar(&c.DestinationDirectory, "destinationpoddirectory", "/backup", "The destination directory to sync the volume data to")
	return fs
}

func SyncVolume(syncVolumeConfig *SyncVolumeFlags) error {

	fmt.Println("syncVolumeConfig", syncVolumeConfig)

	_, getK8sClusterConfigError := k8sutils.GetK8sClusterConfig()

	if getK8sClusterConfigError != nil {
		return getK8sClusterConfigError
	}

	// get status of persistent source volume using the k8sConfig
	// pVolume, getPVolumeError := k8sConfig.ClientSet.CoreV1().PersistentVolumeClaims(syncVolumeConfig.SourceVolumeNamespace).Get(context.TODO(), syncVolumeConfig.SourceVolumeName, metav1.GetOptions{})

	// if getPVolumeError != nil {
	// 	return getPVolumeError
	// }

	// // get status of pVolumes
	// fmt.Printf("pVolume: %s\n", &pVolume.Status)

	return nil
}
