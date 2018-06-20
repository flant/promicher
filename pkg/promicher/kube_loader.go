package promicher

import (
	"encoding/json"
	"fmt"
	"github.com/flant/promicher/pkg/kube"
	"github.com/romana/rlog"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubeResourceData struct {
	Labels      map[string]string
	Annotations map[string]string
}

func (data *KubeResourceData) String() string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("failed to dump json of kube data: %s", err))
	}

	return string(bytes)
}

func LoadKubeResourceData(
	kube *kube.Kube,
	namespace, kind, resourceName string,
	labelsPatterns []string,
	annotationsPatterns []string,
) (*KubeResourceData, error) {
	switch kind {
	case "Pod":
		return LoadPodData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "Deployment":
		return LoadDeploymentData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "ReplicaSet":
		return LoadReplicasetData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "StatefulSet":
		return LoadStatefulsetData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "DaemonSet":
		return LoadDaemonsetData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "Job":
		return LoadJobData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "CronJob":
		return LoadCronJobData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "PersistentVolumeClaim":
		return LoadPersistentVolumeClaimData(kube, namespace, resourceName, labelsPatterns, annotationsPatterns)
	case "Namespace":
		return LoadNamespaceData(kube, namespace, labelsPatterns, annotationsPatterns)
	}

	rlog.Warnf("Unsupported kind '%s' for kube resource '%s/%s' info loader: ignoring resource data", kind, namespace, resourceName)

	return &KubeResourceData{}, nil
}

func LoadOwnerResourcesData(kube *kube.Kube, namespace string, ownerReferences []meta_v1.OwnerReference, labelsPatterns, annotationsPatterns []string) (*KubeResourceData, error) {
	res := &KubeResourceData{}

	for _, ownerRef := range ownerReferences {
		ownerResourceData, err := LoadKubeResourceData(kube, namespace, ownerRef.Kind, ownerRef.Name, labelsPatterns, annotationsPatterns)
		if err != nil {
			return nil, err
		}

		res.Labels = MergeDataMap(res.Labels, ownerResourceData.Labels)
		res.Annotations = MergeDataMap(res.Annotations, ownerResourceData.Annotations)
	}

	return res, nil
}

func LoadObjectData(kube *kube.Kube, namespace string, obj *meta_v1.ObjectMeta, labelsPatterns, annotationsPatterns []string) (*KubeResourceData, error) {
	res, err := MakeObjectData(obj, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	ownersData, err := LoadOwnerResourcesData(kube, namespace, obj.OwnerReferences, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}
	res.Labels = MergeDataMap(res.Labels, ownersData.Labels)
	res.Annotations = MergeDataMap(res.Annotations, ownersData.Annotations)

	if namespace != "" {
		namespaceData, err := LoadNamespaceData(kube, namespace, labelsPatterns, annotationsPatterns)
		if err != nil {
			return nil, err
		}
		res.Labels = MergeDataMap(res.Labels, namespaceData.Labels)
		res.Annotations = MergeDataMap(res.Annotations, namespaceData.Annotations)
	}

	return res, nil
}

func MakeObjectData(obj *meta_v1.ObjectMeta, labelsPatterns, annotationsPatterns []string) (*KubeResourceData, error) {
	res := &KubeResourceData{}

	labels, err := SelectData(obj.Labels, labelsPatterns)
	if err != nil {
		return nil, err
	}
	res.Labels = MergeDataMap(res.Labels, labels)

	annotations, err := SelectData(obj.Annotations, annotationsPatterns)
	if err != nil {
		return nil, err
	}
	res.Annotations = MergeDataMap(res.Annotations, annotations)

	return res, nil
}

func LoadNamespaceData(kube *kube.Kube, resourceName string, labelsPatterns, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.CoreV1().Namespaces().Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube ns/%s %s: %s", resourceName, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, "", &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded ns/%s kube data:\n%s", resourceName, res.String())

	return res, nil
}

func LoadPodData(kube *kube.Kube, namespace, resourceName string, labelsPatterns, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.CoreV1().Pods(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube pod/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded pod/%s kube data from ns/%s:\n%s", resourceName, namespace, res.String())

	return res, nil
}

func LoadDeploymentData(kube *kube.Kube, namespace, resourceName string, labelsPatterns []string, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.AppsV1().Deployments(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube deployment/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded deployment/%s ns/%s kube data:\n%s", resourceName, namespace, res.String())

	return res, nil
}

func LoadReplicasetData(kube *kube.Kube, namespace, resourceName string, labelsPatterns []string, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.AppsV1().ReplicaSets(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube replicaset/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded replicaset/%s kube data from ns/%s:\n%s", resourceName, namespace, res.String())

	return res, nil
}

func LoadStatefulsetData(kube *kube.Kube, namespace, resourceName string, labelsPatterns []string, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.AppsV1().StatefulSets(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube statefulset/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded statefulset/%s kube data from ns/%s:\n%s", resourceName, namespace, res.String())

	return res, nil
}

func LoadDaemonsetData(kube *kube.Kube, namespace, resourceName string, labelsPatterns []string, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.AppsV1().DaemonSets(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube daemonset/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded daemonset/%s kube data from ns/%s:\n%s", resourceName, namespace, res.String())

	return res, nil
}

func LoadJobData(kube *kube.Kube, namespace, resourceName string, labelsPatterns []string, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.BatchV1().Jobs(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube job/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded job/%s kube data from ns/%s:\n%s", resourceName, namespace, res.String())

	return res, nil
}

func LoadCronJobData(kube *kube.Kube, namespace, resourceName string, labelsPatterns []string, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.BatchV1beta1().CronJobs(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube cronjob/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded cronjob/%s kube data from ns/%s:\n%s", resourceName, namespace, res.String())

	return res, nil
}

func LoadPersistentVolumeClaimData(kube *kube.Kube, namespace, resourceName string, labelsPatterns []string, annotationsPatterns []string) (*KubeResourceData, error) {
	resource, err := kube.Client.CoreV1().PersistentVolumeClaims(namespace).Get(resourceName, meta_v1.GetOptions{})
	if err != nil {
		rlog.Errorf("error fetching kube persistentvolumeclaim/%s from ns/%s: %s", resourceName, namespace, err)
		return nil, nil
	}

	res, err := LoadObjectData(kube, namespace, &resource.ObjectMeta, labelsPatterns, annotationsPatterns)
	if err != nil {
		return nil, err
	}

	rlog.Debugf("Loaded persistentvolumeclaim/%s kube data from ns/%s:\n%s", resourceName, namespace, res.String())

	return res, nil
}
