package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"

	"github.com/okteto-community/delete-unused-dev-volumes/app/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const oktetoKubeconfigCommand = "okteto kubeconfig"

func main() {
	ctx := context.Background()
	token := os.Getenv("OKTETO_TOKEN")
	oktetoURL := os.Getenv("OKTETO_URL")

	logLevel := &slog.LevelVar{} // INFO
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	if token == "" || oktetoURL == "" {
		logger.Error("OKTETO_TOKEN and OKTETO_URL environment variables are required")
		os.Exit(1)
	}

	u, err := url.Parse(oktetoURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Invalid OKTETO_URL %s", err))
		os.Exit(1)
	}

	nsList, err := api.GetNamespaces(u.Host, token, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("There was an error requesting the namespaces: %s", err))
		os.Exit(1)
	}

	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		logger.Error(fmt.Sprintf("There was an error creating a temporary directory: %s", err))
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := fmt.Sprintf("%s/.kube/config", tempDir)
	_ = os.Setenv("KUBECONFIG", kubeconfigPath)

	output, err := createKubeconfig()
	if err != nil {
		logger.Error(fmt.Sprintf("There was an error creating the kubeconfig: %s", err))
		os.Exit(1)
	}
	logger.Info(output)

	clientset, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		logger.Error(fmt.Sprintf("There was an error creating the Kubernetes client: %s", err))
		os.Exit(1)
	}
	for _, ns := range nsList {
		logger.Info(fmt.Sprintf("Checking namespace '%s'", ns.Name))

		// We retrieve all the PersistentVolumeClaims mounted in pods in the namespace
		mountedPVCs, err := getMountedPVCs(ctx, clientset, ns.Name)
		if err != nil {
			logger.Error(fmt.Sprintf("Skipping ns %q because there was an error checking PVCs for namespace: %s", ns.Name, err))
			logger.Info("-----------------------------------------------")
			continue
		}

		// We retrieve all the PersistentVolumeClaims created by Okteto for development containers in the namespace
		devPVCs, err := getOktetoDevPVCs(ctx, clientset, ns.Name)
		if err != nil {
			logger.Error(fmt.Sprintf("Skipping ns %q because there was an error checking dev PVCs for namespace: %s", ns.Name, err))
			logger.Info("-----------------------------------------------")
			continue
		}

		if len(devPVCs) == 0 {
			logger.Info(fmt.Sprintf("Skipping ns %q because there are no dev PVCs", ns.Name))
		}

		// For each dev PVC, we delete it if it is not mounted in any pod
		for _, devPVC := range devPVCs {
			if _, ok := mountedPVCs[devPVC]; ok {
				logger.Info(fmt.Sprintf("Skipping PVC %q in namespace %q because it is mounted in a pod", devPVC, ns.Name))
				continue
			}

			if err := deletePVC(ctx, clientset, ns.Name, devPVC); err != nil {
				logger.Error(fmt.Sprintf("Error deleting PVC %q in namespace %q: %s", devPVC, ns.Name, err))
			} else {
				logger.Info(fmt.Sprintf("Deleted PVC %q in namespace %q", devPVC, ns.Name))
			}
		}

		logger.Info("-----------------------------------------------")
	}
}

// deletePVC deletes the PersistentVolumeClaim with the given name in the given namespace
func deletePVC(ctx context.Context, clientset *kubernetes.Clientset, namespace, pvcName string) error {
	err := clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// getOktetoDevPVCs returns the names of the PersistentVolumeClaims created by Okteto for development containers in the given namespace
func getOktetoDevPVCs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]string, error) {
	labelSelector := fmt.Sprintf("dev.okteto.com=true")
	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var devPVCs []string
	for _, pvc := range pvcs.Items {
		devPVCs = append(devPVCs, pvc.Name)
	}

	return devPVCs, nil
}

// getMountedPVCs returns a map of PersistentVolumeClaims mounted in pods in the given namespace
func getMountedPVCs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (map[string]bool, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	mountedPVCs := make(map[string]bool)
	for _, pod := range pods.Items {
		if len(pod.Spec.Volumes) == 0 {
			continue
		}

		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim == nil {
				continue
			}

			mountedPVCs[volume.PersistentVolumeClaim.ClaimName] = true
		}
	}

	return mountedPVCs, nil
}

// createKubeconfig executes the Okteto CLI command to set the kubeconfig to talk with Okteto's cluster
func createKubeconfig() (string, error) {
	cmd := exec.Command("bash", "-c", oktetoKubeconfigCommand)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// getKubernetesClient creates a kubernetes client with the kubeconfig in the server
func getKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("error building k8s config from flags: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
