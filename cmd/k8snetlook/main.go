package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sarun87/k8snetlook/k8snetlook"
)

var (
	podCmd       *flag.FlagSet // Sub-command for pod debugging
	hostOnlyCmd  *flag.FlagSet // Sub-command for host debugging
	podDebugging bool          // Variable to hold debug mode
)

func init() {
	// Pod debugging flags
	podCmd = flag.NewFlagSet("pod", flag.ExitOnError)
	podCmd.StringVar(&k8snetlook.Cfg.SrcPod.Name, "srcpodname", "", "Name of source Pod to debug")
	podCmd.StringVar(&k8snetlook.Cfg.SrcPod.Namespace, "srcpodns", "", "Namespace to which the Pod belongs")
	podCmd.StringVar(&k8snetlook.Cfg.DstPod.Name, "dstpodname", "", "Name of destination Pod to connect")
	podCmd.StringVar(&k8snetlook.Cfg.DstPod.Namespace, "dstpodns", "", "Namespace to which the Pod belongs")
	podCmd.StringVar(&k8snetlook.Cfg.DstSvc.Name, "dstsvcname", "", "Name of detination Service to debug")
	podCmd.StringVar(&k8snetlook.Cfg.DstSvc.Namespace, "dstsvcns", "", "Namespace to which the Pod belongs")
	podCmd.StringVar(&k8snetlook.Cfg.ExternalIP, "externalip", "", "External IP to test egress traffic flow")
	podCmd.StringVar(&k8snetlook.Cfg.KubeconfigPath, "config", os.Getenv("KUBECONFIG"), "Path to Kubeconfig")

	hostOnlyCmd = flag.NewFlagSet("host", flag.ExitOnError)
	hostOnlyCmd.StringVar(&k8snetlook.Cfg.KubeconfigPath, "config", os.Getenv("KUBECONFIG"), "Path to Kubeconfig")

}

func printUsage() {
	fmt.Println("")
	fmt.Println("usage: k8snetlook subcommand [sub-command-options] [-config path-to-kube-config] ")
	fmt.Println("")
	fmt.Println("valid subcommands")
	fmt.Println("  pod       Debug Pod & host networking")
	fmt.Println("  host      Debug host networking only")
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("'host' or 'pod' subcommand expected")
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "pod":
		podDebugging = true
		podCmd.Parse(os.Args[2:])
	case "host":
		hostOnlyCmd.Parse(os.Args[2:])
	default:
		fmt.Println("'host' or 'pod' subcommand expected")
		printUsage()
		os.Exit(1)
	}

	validateArgs()

	k8snetlook.InitKubeClient(k8snetlook.Cfg.KubeconfigPath)
	k8snetlook.InitK8sInfo()
	defer k8snetlook.Cleanup()

	k8snetlook.RunHostChecks()
	if podDebugging == true {
		k8snetlook.RunPodChecks()
	}
}

func validateArgs() {
	fmt.Println("")
	if k8snetlook.Cfg.KubeconfigPath == "" {
		fmt.Println("error: KUBECONFIG env variable not set. Please pass kubeconfig using -config")
		printUsage()
		os.Exit(1)
	}
	if podDebugging && (k8snetlook.Cfg.SrcPod.Name == "" || k8snetlook.Cfg.SrcPod.Namespace == "") {
		fmt.Printf("error: srcpodname flag and srcpodns required for pod debugging\n\n")
		podCmd.Usage()
		os.Exit(1)
	}
}
