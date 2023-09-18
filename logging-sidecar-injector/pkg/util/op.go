package util

import (
	"context"
	"fmt"
	"github.com/bwei/logging-sidecar-injector/pkg/apis/dbs/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"log"
)

type LogDetails struct {
	FileName string
	Name     string
}

type Log struct {
	Flist     []LogDetails
	V         corev1.Volume
	Path      string
	Container string
	Name      string
}

func RestClient() *rest.RESTClient {
	var config *rest.Config
	var err error
	log.Printf("using in-cluster configuration")
	config, err = rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	v1alpha1.AddToScheme(scheme.Scheme)
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupName, Version: v1alpha1.GroupVersion}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	RestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		panic(err)
		return nil
	}
	return RestClient
}

func Fetch(client *rest.RESTClient, namespace string, name string) *v1alpha1.LogInfo {
	result := &v1alpha1.LogInfo{}
	client.Get().Resource("loginfos").Namespace(namespace).Name(name).Do(context.TODO()).Into(result)
	return result
}

func CreateVolume(fname string) corev1.Volume {
	v := corev1.Volume{Name: fname}
	v.VolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
	return v
}

func PatchDeploy(log Log, deploy *appsv1.Deployment, imagename string) {
	vm := corev1.VolumeMount{}
	vm.Name = log.V.Name
	vm.MountPath = log.Path
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, log.V)
	for i, c := range deploy.Spec.Template.Spec.Containers {
		if c.Name == log.Container {
			deploy.Spec.Template.Spec.Containers[i].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[i].VolumeMounts, vm)
		}
	}
	for _, f := range log.Flist {
		sidecar := CreateSidecarContainer(f.FileName, log.Name, f.Name, imagename)
		deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, sidecar)
	}

}

func CreateSidecarContainer(filename string, fname string, cname string, imagename string) corev1.Container {
	vm := corev1.VolumeMount{}
	vm.Name = fname
	vm.MountPath = "/var/log"
	resource := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    resource.MustParse("20m"),
			"memory": resource.MustParse("20Mi"),
		},
		Requests: corev1.ResourceList{
			"cpu":    resource.MustParse("10m"),
			"memory": resource.MustParse("10Mi"),
		},
	}
	con := corev1.Container{Name: fmt.Sprintf("%s-log-sidecar", cname), ImagePullPolicy: "Always", Image: imagename, Args: []string{"/bin/sh", "-c", fmt.Sprintf("tail -n+1 -F /var/log/%s", filename)}, VolumeMounts: []corev1.VolumeMount{vm}, Resources: resource}
	return con
}
