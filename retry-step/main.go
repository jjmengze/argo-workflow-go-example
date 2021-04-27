package main

import (
	"argo-workflow-goexample/utils"
	"context"
	"fmt"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	intstrutil "github.com/argoproj/argo-workflows/v3/util/intstr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags(utils.MasterURL, utils.KubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	//var parameter = []wfv1.Parameter{{Name: "cmd"}}
	var parameterA = []wfv1.Parameter{{Name: "cmd", Value: wfv1.AnyStringPtr("import random; import sys; exit_code = random.choice([0, 1, 1, 1]);  print(exit_code);sys.exit(exit_code)")}}
	var parameterAA = []wfv1.Parameter{{Name: "cmd", Value: wfv1.AnyStringPtr("import random; import sys; exit_code = random.choice([0, 1, 1, 1]);  print(exit_code);sys.exit(exit_code)")}}

	var stepsdWorkflow = wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "retry-backoff",
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "step-by-step",
			Templates: []wfv1.Template{
				{
					Name: "retry-backoff",
					Container: &corev1.Container{
						Image:   "python:alpine3.6",
						Command: []string{"python", "-c"},
						Args: []string{
							"{{inputs.parameters.cmd}}",
						},
					},
					RetryStrategy: &wfv1.RetryStrategy{
						Limit:       intstrutil.ParsePtr("5"),
						RetryPolicy: utils.RetryPolicyAlways,
						Backoff: &wfv1.Backoff{
							Duration:    "2s",
							Factor:      intstrutil.ParsePtr("2"),
							MaxDuration: "1m",
						},
						//Affinity:    nil, not used
					},
					Inputs: wfv1.Inputs{
						Parameters: parameterA,
					},
				},
				{
					Name: "step-by-step",
					Steps: []wfv1.ParallelSteps{
						{
							Steps: []wfv1.WorkflowStep{
								{
									Name:     "Step-1",
									Template: "retry-backoff",
									Arguments: wfv1.Arguments{
										Parameters: parameterA,
									},
								},
								{
									Name:     "Step-1A",
									Template: "retry-backoff",
									Arguments: wfv1.Arguments{
										Parameters: parameterAA,
									},
								},
							},
						},
					},
				},
			},
		}}

	// create the workflow client
	wfClient := wfclientset.NewForConfigOrDie(config).ArgoprojV1alpha1().Workflows(utils.NameSpace)
	// submit the hello world workflow
	ctx := context.Background()
	createdWf, err := wfClient.Create(ctx, &stepsdWorkflow, metav1.CreateOptions{})
	if err != nil {
		klog.Fatalf("Error creating argo workflow: %s", err.Error())
	}
	klog.Infof("Workflow %s submitted\n", createdWf.Name)
	// wait for the workflow to complete
	fieldSelector := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", createdWf.Name))
	watchIf, err := wfClient.Watch(ctx, metav1.ListOptions{FieldSelector: fieldSelector.String(), TimeoutSeconds: pointer.Int64Ptr(180)})
	defer watchIf.Stop()
	for next := range watchIf.ResultChan() {
		wf, ok := next.Object.(*wfv1.Workflow)
		if !ok {
			continue
		}
		if !wf.Status.FinishedAt.IsZero() {
			fmt.Printf("Workflow %s %s at %v. Message: %s.\n", wf.Name, wf.Status.Phase, wf.Status.FinishedAt, wf.Status.Message)
			break
		}
	}
}
