package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	srcPodName   string        // Variable to store name of source pod
	dstPodName   string        // Variable to store name of destination pod
	dstSvcName   string        // Variable to store name of destination service
	kubeconfig   string        // Variable to store path to kubeconfig
	externalIP   string        // Variable to store external IP destination
	podCmd       *flag.FlagSet // Sub-command for pod debugging
	hostCmd      *flag.FlagSet // Sub-command for host debugging
	podDebugging bool          // Variable to hold debug mode
)

func init() {
	// Pod debugging flags
	podCmd = flag.NewFlagSet("pod", flag.ExitOnError)
	podCmd.StringVar(&srcPodName, "srcpod", "", "Name of source Pod to debug")
	podCmd.StringVar(&dstPodName, "dstpod", "", "Name of destination Pod to connect")
	podCmd.StringVar(&dstSvcName, "dstsvc", "kubernetes", "Name of detination Service to debug")
	podCmd.StringVar(&externalIP, "externalip", "8.8.8.8", "External IP to test egress traffic flow")
	podCmd.StringVar(&kubeconfig, "config", os.Getenv("KUBECONFIG"), "Path to Kubeconfig")

	// Host debugging flags
	hostCmd = flag.NewFlagSet("host", flag.ExitOnError)
	hostCmd.StringVar(&externalIP, "externalip", "8.8.8.8", "External IP to test egress traffic flow")
	hostCmd.StringVar(&kubeconfig, "config", os.Getenv("KUBECONFIG"), "Path to Kubeconfig")
}

func printUsage() {
	fmt.Println("")
	fmt.Println("usage: k8snetlook subcommand [sub-command-options]")
	fmt.Println("")
	fmt.Println("valid subcommands")
	fmt.Println("  pod    Debug Pod networking")
	fmt.Println("  host   Debug generic networking from the host")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("error: expected 'pod' or 'host' subcommand")
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "pod":
		podDebugging = true
		podCmd.Parse(os.Args[2:])
	case "host":
		hostCmd.Parse(os.Args[2:])
	default:
		fmt.Println("error: expected 'pod' or 'host' subcommand")
		printUsage()
		os.Exit(1)
	}

	validateArgs()

	// Start debugging process

}

func validateArgs() {
	exitProgram := false
	fmt.Println("")
	if kubeconfig == "" {
		fmt.Println("error: KUBECONFIG env variable not set. Please pass kubeconfig using -config")
		exitProgram = true
	}
	if podDebugging && srcPodName == "" {
		fmt.Printf("error: srcpod flag required for pod debugging\n\n")
		podCmd.Usage()
		exitProgram = true
	}
	if exitProgram {
		os.Exit(1)
	}
}
