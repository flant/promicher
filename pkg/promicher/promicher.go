package promicher

import (
	"github.com/flant/promicher/pkg/kube"
	"github.com/romana/rlog"
)

type Promicher struct {
	Kube        *kube.Kube
	AlertsCache map[string]Alert
	Labels      []string
	Annotations []string
}

func NewPromicher(kube *kube.Kube, labels, annotations []string) *Promicher {
	return &Promicher{
		Kube:        kube,
		AlertsCache: make(map[string]Alert),
		Labels:      labels,
		Annotations: annotations,
	}
}

func (promicher *Promicher) ProcessAlert(alert Alert) (Alert, error) {
	resource := alert.KubeTargetResourceInfo()
	if resource == nil {
		return alert, nil
	}

	if !alert.EndsAt.IsZero() {
		if cachedAlert, hasKey := promicher.AlertsCache[resource.CacheId()]; hasKey {
			rlog.Debugf("Cache hit for resource '%s':\n%s", resource.CacheId(), cachedAlert.String())

			return cachedAlert, nil
		}
	}

	data, err := LoadKubeResourceData(promicher.Kube, resource.Namespace, resource.Kind, resource.Name, promicher.Labels, promicher.Annotations)
	if err != nil {
		return Alert{}, err
	}

	if data == nil {
		if cachedAlert, hasKey := promicher.AlertsCache[resource.CacheId()]; hasKey {
			rlog.Debugf("Cache hit for resource '%s':\n%s", resource.CacheId(), cachedAlert.String())

			return cachedAlert, nil
		}
	} else {
		alert.Labels = MergeDataMap(alert.Labels, data.Labels)
		alert.Annotations = MergeDataMap(alert.Annotations, data.Annotations)

		promicher.AlertsCache[resource.CacheId()] = alert

		rlog.Debugf("Cache updated for alert '%s':\n%s", resource.CacheId(), alert.String())
	}

	return alert, nil
}

func (promicher *Promicher) ProcessData(dataBytes []byte) ([]byte, error) {
	alerts, err := ParseAlerts(dataBytes)
	if err != nil {
		return nil, err
	}

	res := make([]Alert, 0)

	for _, alert := range alerts {
		newAlert, err := promicher.ProcessAlert(alert)
		if err != nil {
			return nil, err
		}
		res = append(res, newAlert)
	}

	return DumpAlerts(res)
}
