package main

import (
	"argo-workflow-goexample/utils"
	"context"
	"fmt"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

var parameter = []wfv1.Parameter{{Name: "message"}}
var parameterA = []wfv1.Parameter{{Name: "message", Value: wfv1.AnyStringPtr("Hello-A")}}
var parameterAA = []wfv1.Parameter{{Name: "message", Value: wfv1.AnyStringPtr("Hello-AA")}}
var parameterAAA = []wfv1.Parameter{{Name: "message", Value: wfv1.AnyStringPtr("Hello-AAA")}}
var parameterAAAA = []wfv1.Parameter{{Name: "message", Value: wfv1.AnyStringPtr("Hello-AAAA")}}
var parameterAAAAA = []wfv1.Parameter{{Name: "message", Value: wfv1.AnyStringPtr("Hello-AAAAA")}}

var parameterB = []wfv1.Parameter{{Name: "message", Value: wfv1.AnyStringPtr("Hello-B")}}

func main() {
	config, err := clientcmd.BuildConfigFromFlags(utils.MasterURL, utils.KubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	var stepsdWorkflow = wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "hello-word-",
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "step-by-step",
			Templates: []wfv1.Template{
				{
					Name: "whalesay",
					Container: &corev1.Container{
						Image:   "docker/whalesay:latest",
						Command: []string{"cowsay"},
						Args: []string{
							"{{inputs.parameters.message}}",
						},
					},
					Inputs: wfv1.Inputs{
						Parameters: []wfv1.Parameter{{Name: "message"}},
					},
				},
				{
					Name: "step-by-step",
					Steps: []wfv1.ParallelSteps{
						{
							Steps: []wfv1.WorkflowStep{
								{
									Name:     "Step-1",
									Template: "whalesay",
									Arguments: wfv1.Arguments{
										Parameters: parameterA,
									},
								},
								{
									Name:     "Step-1A",
									Template: "whalesay",
									Arguments: wfv1.Arguments{
										Parameters: parameterAA,
									},
								},
								{
									Name:     "Step-1AA",
									Template: "whalesay",
									Arguments: wfv1.Arguments{
										Parameters: parameterAAA,
									},
								},
								{
									Name:     "Step-1AAA",
									Template: "whalesay",
									Arguments: wfv1.Arguments{
										Parameters: parameterAAAA,
									},
								},
								{
									Name:     "Step-1AAAA",
									Template: "whalesay",
									Arguments: wfv1.Arguments{
										Parameters: parameterAAAAA,
									},
								},
							},
						},
						{
							Steps: []wfv1.WorkflowStep{
								{
									Name:     "Step-2",
									Template: "whalesay",
									Arguments: wfv1.Arguments{
										Parameters: []wfv1.Parameter{{Name: "message", Value: wfv1.AnyStringPtr("Hello-B")}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

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
