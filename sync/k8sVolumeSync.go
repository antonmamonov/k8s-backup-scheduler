package sync

import (
	"flag"
	"fmt"
	"os/exec"
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

	// build kubectl cp command
	// kubectl cp <namespace>/<pod-name>:/path/to/remote/file /path/to/local/file
	kubectlCpString := fmt.Sprintf("/usr/local/bin/kubectl cp %s/%s:%s %s", syncVolumeConfig.SourcePodNamespace, syncVolumeConfig.SourcePodName, syncVolumeConfig.SourcePodDirectory, syncVolumeConfig.DestinationDirectory)

	fmt.Println("kubectlCpString", kubectlCpString)

	// exec kubectl cp command
	execCommandError := exec.Command("/usr/local/bin/kubectl", "cp", fmt.Sprintf("%s/%s:%s", syncVolumeConfig.SourcePodNamespace, syncVolumeConfig.SourcePodName, syncVolumeConfig.SourcePodDirectory), syncVolumeConfig.DestinationDirectory).Run()

	return execCommandError
}
