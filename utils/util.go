package utils

import "flag"

func init() {
	flag.StringVar(&KubeConfig, "kubeconfig", "", "Path to a KubeConfig. Only required if out-of-cluster.")
	flag.StringVar(&MasterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in KubeConfig. Only required if out-of-cluster.")
	flag.StringVar(&NameSpace, "mamespace", ArgoNs, "The work is running in specific NameSpace.")
	flag.Parse()

}

const (
	ArgoNs            = "argo"
	RetryPolicyAlways = "Always"
)

var (
	MasterURL  string
	KubeConfig string
	NameSpace  string
)
