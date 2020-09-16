package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sarun87/k8snetlook/k8snetlook"
	log "github.com/sarun87/k8snetlook/logutil"
)

var (
	podCmd       *flag.FlagSet // Sub-command for pod debugging
	hostOnlyCmd  *flag.FlagSet // Sub-command for host debugging
	podDebugging bool          // Variable to hold debug mode
	debugLogging bool          // Enable debug logging
	silent       bool          // Output only errors
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
	podCmd.BoolVar(&debugLogging, "debug", false, "Enable debug logging to stdout")
	podCmd.BoolVar(&silent, "silent", false, "Output only errors to stdout")

	hostOnlyCmd = flag.NewFlagSet("host", flag.ExitOnError)
	hostOnlyCmd.StringVar(&k8snetlook.Cfg.KubeconfigPath, "config", os.Getenv("KUBECONFIG"), "Path to Kubeconfig")
	hostOnlyCmd.BoolVar(&debugLogging, "debug", false, "Enable debug logging to stdout")
	hostOnlyCmd.BoolVar(&silent, "silent", false, "Output only errors to stdout. Return result as json")

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

	logLevel := log.INFO
	if debugLogging {
		logLevel = log.DEBUG
	}
	if silent {
		logLevel = log.ERROR
	}

	log.SetLogLevel(logLevel)

	if err := k8snetlook.Init(k8snetlook.Cfg.KubeconfigPath); err != nil {
		log.Error("Unable to initialize k8snetlook\n")
		log.Error("%v\n", err)
		return
	}
	defer k8snetlook.Cleanup()

	k8snetlook.RunHostChecks()
	if podDebugging == true {
		k8snetlook.RunPodChecks()
	}
	if silent {
		fmt.Printf("%s", k8snetlook.GetReportJSON())
		return
	}
	k8snetlook.PrintReport()
}

func validateArgs() {
	fmt.Println("")
	if podDebugging && (k8snetlook.Cfg.SrcPod.Name == "" || k8snetlook.Cfg.SrcPod.Namespace == "") {
		fmt.Printf("error: srcpodname flag and srcpodns required for pod debugging\n\n")
		podCmd.Usage()
		os.Exit(1)
	}
}
