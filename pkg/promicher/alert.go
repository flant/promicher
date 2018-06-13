package promicher

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAtRaw string            `json:"startsAt"`
	EndsAtRaw   string            `json:"endsAt"`

	StartsAt time.Time `json:"-"`
	EndsAt   time.Time `json:"-"`

	raw map[string]interface{} `json:",inline"`
}

func (alert *Alert) UnmarshalJSON(b []byte) error {
	var err error

	type rawAlert Alert
	err = json.Unmarshal(b, (*rawAlert)(alert))
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &alert.raw)
	if err != nil {
		return err
	}

	return nil
}

func (alert *Alert) MarshalJSON() ([]byte, error) {
	alert.raw["labels"] = alert.Labels
	alert.raw["annotations"] = alert.Annotations
	alert.raw["startsAt"] = alert.StartsAtRaw
	alert.raw["endsAt"] = alert.EndsAtRaw

	return json.Marshal(alert.raw)
}

type KubeResourceInfo struct {
	Namespace string
	Kind      string
	Name      string
}

func (resource *KubeResourceInfo) CacheId() string {
	return fmt.Sprintf("ns/%s %s/%s", resource.Namespace, strings.ToLower(resource.Kind), resource.Name)
}

func (alert *Alert) String() string {
	alertBytes, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("failed to dump json of prometheus alert: %s", err))
	}

	return string(alertBytes)
}

func (alert *Alert) KubeTargetResourceInfo() *KubeResourceInfo {
	if ns, hasKey := alert.Labels["namespace"]; hasKey {
		for _, kind := range []string{
			"Pod",
			"Deployment",
			"StatefulSet",
			"DaemonSet",
			"Job",
			"CronJob",
			"PersistentVolumeClaim",
			"PersistentVolume",
		} {
			if name, hasKey := alert.Labels[strings.ToLower(kind)]; hasKey {
				return &KubeResourceInfo{
					Namespace: ns,
					Kind:      kind,
					Name:      name,
				}
			}
		}

		return &KubeResourceInfo{
			Namespace: "",
			Kind:      "Namespace",
			Name:      ns,
		}
	}

	return nil
}

func ParseAlerts(data []byte) ([]Alert, error) {
	var res []Alert
	var err error

	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	for i := range res {
		alert := &res[i]

		alert.StartsAt, err = time.Parse(time.RFC3339, alert.StartsAtRaw)
		if err != nil {
			return nil, fmt.Errorf("Bad alert `startsAt` field data \"%s\": %s", alert.StartsAtRaw, err)
		}

		alert.EndsAt, err = time.Parse(time.RFC3339, alert.EndsAtRaw)
		if err != nil {
			return nil, fmt.Errorf("Bad alert `endsAt` field data \"%s\": %s", alert.EndsAtRaw, err)
		}
	}

	return res, nil
}

func DumpAlerts(alerts []Alert) ([]byte, error) {
	data, err := json.Marshal(alerts)
	if err != nil {
		return nil, err
	}
	return data, nil
}
