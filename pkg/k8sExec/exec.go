package k8sExec

import (
	"bytes"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type K8sExec struct {
	ClientSet  kubernetes.Interface
	RestConfig *rest.Config

	PodName       string
	ContainerName string
	Namespace     string
}

func New(clientSet kubernetes.Interface, restConfig *rest.Config, podName, containerName, namespace string) *K8sExec {
	return &K8sExec{
		ClientSet:     clientSet,
		RestConfig:    restConfig,
		PodName:       podName,
		ContainerName: containerName,
		Namespace:     namespace,
	}
}

func (k8s *K8sExec) Exec(command []string) ([]byte, []byte, error) {
	req := k8s.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(k8s.PodName).
		Namespace(k8s.Namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: k8s.ContainerName,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	log.Infof("Request URL: %s", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(k8s.RestConfig, "POST", req.URL())
	if err != nil {
		log.Errorf("Failed to exec:%v", err)
		return []byte{}, []byte{}, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		log.Errorf("Faile to get result:%v", err)
		return []byte{}, []byte{}, err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

func (k8s *K8sExec) PutToPod(content, remoteFile string) error {
	req := k8s.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(k8s.PodName).
		Namespace(k8s.Namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: k8s.ContainerName,
		Command:   []string{"sh", "-c", fmt.Sprintf("cat > %s", remoteFile)},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	log.Infof("Request URL: %s", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(k8s.RestConfig, "POST", req.URL())
	if err != nil {
		log.Errorf("Failed to exec:%v", err)
		return err
	}

	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()
		buf := bytes.NewBufferString(content)
		io.Copy(writer, buf)
	}()

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		log.Errorf("stdout:%s\nstderr:%s", stdout.Bytes(), stderr.Bytes())
		log.Errorf("Failed to get result:%v", err)
		return err
	}
	log.Debugf("stdout:%s\nstderr:%s", stdout.Bytes(), stderr.Bytes())
	return nil
}

func (k8s *K8sExec) DownloadFromPod(remoteFile, localFile string) error {
	req := k8s.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(k8s.PodName).
		Namespace(k8s.Namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: k8s.ContainerName,
		Command:   []string{"sh", "-c", fmt.Sprintf("cat %s", remoteFile)},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	log.Infof("Request URL: %s", req.URL().String())

	exec, err := remotecommand.NewSPDYExecutor(k8s.RestConfig, "POST", req.URL())
	if err != nil {
		log.Errorf("Failed to exec:%v", err)
		return err
	}

	reader, writer := io.Pipe()
	var stderr bytes.Buffer

	go func() {
		defer writer.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  nil,
			Stdout: writer,
			Stderr: &stderr,
		})
		if err != nil {
			log.Errorf("stderr:%s", stderr.Bytes())
			log.Errorf("Failed to get result:%v", err)
			return
		}
	}()

	file, err := os.Create(localFile)
	if err != nil {
		log.Errorf("Failed to create file:%v", err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		log.Errorf("Failed to save to file:%v", err)
		return err
	}
	reader.Close()
	return nil
}
