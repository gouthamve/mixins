package common

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/resource"
)

const (
	PrometheusDSType = "prometheus"
	PrometheusDSUid  = "prom"

	LokiDSType = "loki"
	LokiDSUid  = "loki"
)

func DashboardManifest(folderUid string, dash dashboard.Dashboard) resource.Manifest {
	return resource.Manifest{
		ApiVersion: "dashboard.grafana.app/v1beta1",
		Kind:       "Dashboard",
		Metadata: resource.Metadata{
			Annotations: map[string]string{
				"grafana.app/folder": folderUid,
			},
			Name: *dash.Uid,
		},
		Spec: dash,
	}
}

func NewPrometheusDSSelector() *dashboard.DatasourceVariableBuilder {
	return dashboard.NewDatasourceVariableBuilder("prom").
		Name(PrometheusDSUid).
		Type(PrometheusDSType)
}

func NewLokiDSSelector() *dashboard.DatasourceVariableBuilder {
	return dashboard.NewDatasourceVariableBuilder("loki").
		Name(LokiDSUid).
		Type(LokiDSType)
}

func HTTPRedReqsOverrides() []cog.Builder[dashboard.DashboardFieldConfigSourceOverrides] {
	overrides := []struct {
		name  string
		color string
	}{
		{"1xx", "#EAB839"},
		{"2xx", "#7EB26D"},
		{"3xx", "#6ED0E0"},
		{"4xx", "#EF843C"},
		{"5xx", "#E24D42"},
	}

	fieldOverrides := make([]cog.Builder[dashboard.DashboardFieldConfigSourceOverrides], 0, len(overrides))
	for _, o := range overrides {
		fieldOverrides = append(fieldOverrides, dashboard.NewDashboardFieldConfigSourceOverridesBuilder().
			Matcher(dashboard.MatcherConfig{
				Id:      "byName",
				Options: o.name,
			}).
			Properties([]dashboard.DynamicConfigValue{
				{
					Id: "color",
					Value: dashboard.FieldColor{
						Mode:       "fixed",
						FixedColor: &o.color,
					},
				},
			}),
		)
	}

	return fieldOverrides
}
