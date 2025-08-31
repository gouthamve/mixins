package dashboard

import (
	"fmt"
	"strings"

	mixincommon "github.com/gouthamve/mixins/common"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/cog/variants"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/logs"
	"github.com/grafana/grafana-foundation-sdk/go/loki"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
	"github.com/grafana/grafana-foundation-sdk/go/table"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

type Config struct {
	LogsQuery string

	ServiceNamespaces []string
	ServiceNames      []string
}

func Build(cfg Config) (dashboard.Dashboard, error) {
	builder := dashboard.NewDashboardBuilder("OTel Application Semantic convention").
		Uid("otel-app-semantic-conventions").
		Tags([]string{"otel", "generated"}).
		Refresh("1m").
		Time("now-60m", "now").
		Timezone(common.TimeZoneBrowser).
		Variables([]cog.Builder[dashboard.VariableModel]{
			mixincommon.NewPrometheusDSSelector(),
			mixincommon.NewLokiDSSelector(),
			serviceNamespaceVar(cfg.ServiceNamespaces),
			serviceNameVar(cfg.ServiceNames),
		}).
		WithRow(dashboard.NewRowBuilder("HTTP RED Overview")).
		WithPanel(rpsOverviewPanel()).
		WithPanel(latencyOverviewPanel()).
		WithRow(dashboard.NewRowBuilder("HTTP Route Details")).
		WithPanel(rpsDetailsPanel())

	if cfg.LogsQuery != "" {
		builder = builder.WithRow(dashboard.NewRowBuilder("Logs")).
			WithPanel(logsTablePanel(cfg.LogsQuery))
	}

	return builder.Build()
}

func rpsOverviewPanel() *timeseries.PanelBuilder {
	return mixincommon.NewTimeseriesPanelBuilder().
		Title("Requests/sec").
		Unit("reqps").
		Min(0).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(httpReqsQuery).
				LegendFormat("{{ status }}"),
		).
		Overrides(mixincommon.HTTPRedReqsOverrides())

}

func latencyOverviewPanel() *timeseries.PanelBuilder {
	return mixincommon.NewTimeseriesPanelBuilder().
		Title("Latency").
		Unit("s").
		Min(0).
		Targets([]cog.Builder[variants.Dataquery]{
			prometheus.NewDataqueryBuilder().
				Expr(p99LatencyQuery).
				LegendFormat("99th percentile"),
			prometheus.NewDataqueryBuilder().
				Expr(p50LatencyQuery).
				LegendFormat("50th percentile"),
			prometheus.NewDataqueryBuilder().
				Expr(avgLatencyQuery).
				LegendFormat("Average"),
		})
}

func rpsDetailsPanel() *table.PanelBuilder {
	return mixincommon.NewTablePanelBuilder().
		Targets([]cog.Builder[variants.Dataquery]{
			prometheus.NewDataqueryBuilder().
				Expr(reqsRateOpsQuery).
				Range().
				RefId(reqsRateOpsQueryId),
			prometheus.NewDataqueryBuilder().
				Expr(reqsRateOpsErrorQuery).
				Range().
				RefId(reqsRateOpsErrorQueryId),
			prometheus.NewDataqueryBuilder().
				Expr(p95LatencyOpsQuery).
				Range().
				RefId(p95LatencyOpsQueryId),
		}).
		GridPos(dashboard.GridPos{
			H: 10,
			W: 24,
		}).
		SortBy([]cog.Builder[common.TableSortByFieldState]{common.NewTableSortByFieldStateBuilder().DisplayName("duration p95").Desc(true)}).
		Transformations([]dashboard.DataTransformerConfig{
			{
				Id: "timeSeriesTable",
				Options: map[string]map[string]any{
					reqsRateOpsQueryId: {
						"timeField": "Time",
					},
					reqsRateOpsErrorQueryId: {
						"timeField": "Time",
					},
					p95LatencyOpsQueryId: {
						"timeField": "Time",
					},
				},
			},
			{
				Id: "joinByField",
				Options: map[string]any{
					"byField": "operation",
					"mode":    "outer",
				},
			},
			{
				Id: "organize",
				Options: map[string]any{
					"excludeByName": map[string]bool{
						"http_request_method 1": true,
						"http_request_method 2": true,
						"http_request_method 3": true,
						"http_route 1":          true,
						"http_route 2":          true,
						"http_route 3":          true,
					},
					"renameByName": map[string]string{
						fmt.Sprintf("Trend #%s", reqsRateOpsQueryId):      "rate",
						fmt.Sprintf("Trend #%s", reqsRateOpsErrorQueryId): "errors",
						fmt.Sprintf("Trend #%s", p95LatencyOpsQueryId):    "duration p95",
					},
					"indexByName": map[string]int{
						"operation":               0,
						"Trend #p95LatencyOps":    1,
						"Trend #reqsRateOps":      2,
						"Trend #reqsRateOpsError": 3,
					},
				},
			},
		}).Overrides([]cog.Builder[dashboard.DashboardFieldConfigSourceOverrides]{
		dashboard.NewDashboardFieldConfigSourceOverridesBuilder().
			Matcher(dashboard.MatcherConfig{
				Id:      "byName",
				Options: "duration p95",
			}).
			Properties([]dashboard.DynamicConfigValue{
				{
					Id:    "unit",
					Value: "s",
				},
				{
					Id:    "color",
					Value: dashboard.FieldColor{Mode: "shades", FixedColor: cog.ToPtr("orange")},
				},
			}),
		dashboard.NewDashboardFieldConfigSourceOverridesBuilder().
			Matcher(dashboard.MatcherConfig{
				Id:      "byName",
				Options: "rate",
			}).
			Properties([]dashboard.DynamicConfigValue{
				{
					Id:    "unit",
					Value: "reqps",
				},
				{
					Id:    "color",
					Value: dashboard.FieldColor{Mode: "shades", FixedColor: cog.ToPtr("green")},
				},
			}),
		dashboard.NewDashboardFieldConfigSourceOverridesBuilder().
			Matcher(dashboard.MatcherConfig{
				Id:      "byName",
				Options: "errors",
			}).
			Properties([]dashboard.DynamicConfigValue{
				{
					Id:    "unit",
					Value: "percentunit",
				},
				{
					Id:    "color",
					Value: dashboard.FieldColor{Mode: "shades", FixedColor: cog.ToPtr("red")},
				},
			}),
	})
}

func logsTablePanel(logsQuery string) *logs.PanelBuilder {
	return mixincommon.NewLogsPanelBuilder().
		WithTarget(
			loki.NewDataqueryBuilder().
				Expr(logsQuery).
				Range(true),
		).
		EnableInfiniteScrolling(true).
		GridPos(dashboard.GridPos{
			H: 10,
			W: 24,
		})

}

func serviceNamespaceVar(options []string) *dashboard.QueryVariableBuilder {
	queryVar := dashboard.NewQueryVariableBuilder("service_namespace").
		Label("Service Namespace").
		Query(dashboard.StringOrMap{String: cog.ToPtr(serviceNamespaceVarQuery)}).
		Datasource(dashboard.DataSourceRef{
			Type: cog.ToPtr(mixincommon.PrometheusDSType),
			Uid:  cog.ToPtr("${" + mixincommon.PrometheusDSUid + "}"),
		}).
		Multi(true).
		IncludeAll(true).
		AllValue(".*").
		Refresh(dashboard.VariableRefreshOnDashboardLoad)

	if len(options) > 0 {
		optionsRegex := fmt.Sprintf("/^(%s)$/", strings.Join(options, "|"))
		queryVar = queryVar.Regex(optionsRegex).AllValue("")
		if len(options) == 1 {
			queryVar = queryVar.Multi(false).IncludeAll(false)
		}
	}

	return queryVar
}

func serviceNameVar(options []string) *dashboard.QueryVariableBuilder {
	queryVar := dashboard.NewQueryVariableBuilder("service_name").
		Label("Service Name").
		Query(dashboard.StringOrMap{
			String: cog.ToPtr(serviceNameVarQuery),
		}).
		Datasource(dashboard.DataSourceRef{
			Type: cog.ToPtr(mixincommon.PrometheusDSType),
			Uid:  cog.ToPtr("${" + mixincommon.PrometheusDSUid + "}"),
		}).
		Multi(true).
		IncludeAll(true).
		AllValue(".+").
		Refresh(dashboard.VariableRefreshOnDashboardLoad)

	if len(options) > 0 {
		optionsRegex := fmt.Sprintf("/^(%s)$/", strings.Join(options, "|"))
		queryVar = queryVar.Regex(optionsRegex).AllValue("")
		if len(options) == 1 {
			queryVar = queryVar.Multi(false).IncludeAll(false)
		}
	}

	return queryVar
}

var (
	serviceNamespaceVarQuery = `label_values(target_info, service_namespace)`
	serviceNameVarQuery      = `label_values(target_info{service_namespace=~"$service_namespace"}, service_name)`

	selectors = `service_namespace=~"$service_namespace", service_name=~"$service_name"`

	httpReqsQuery = fmt.Sprintf(`
sum by (status) (
  label_replace(
    rate(
      http_server_request_duration_seconds_count{%s}[$__rate_interval]
    ),
    "status",
    "${1}xx",
    "http_response_status_code",
    "([0-9]).."
  )
)`, selectors)

	p99LatencyQuery = fmt.Sprintf(`
histogram_quantile(
  0.99,
  sum by (le) (
    rate(
      http_server_request_duration_seconds_bucket{%s}[$__rate_interval]
    )
  )
)`, selectors)

	p50LatencyQuery = fmt.Sprintf(`
histogram_quantile(
	0.50,	
	sum by (le) (
		rate(
			http_server_request_duration_seconds_bucket{%s}[$__rate_interval]
		)
	)
)`, selectors)

	avgLatencyQuery = fmt.Sprintf(`
  sum(
    rate(
      http_server_request_duration_seconds_sum{%s}[$__rate_interval]
    )
  )
/
  sum(
    rate(
      http_server_request_duration_seconds_count{%s}[$__rate_interval]
    )
  )
`, selectors, selectors)

	reqsRateOpsQuery = fmt.Sprintf(`
label_join(
  sum by (http_request_method, http_route) (
    rate(
      http_server_request_duration_seconds_count{%s}[$__rate_interval]
    )
  ),
  "operation",
  " ",
  "http_request_method",
  "http_route"
)`, selectors)
	reqsRateOpsQueryId = "reqsRateOps"

	reqsRateOpsErrorQuery = fmt.Sprintf(`
(label_join(
  sum by (http_request_method, http_route) (
    rate(
      http_server_request_duration_seconds_count{http_response_status_code=~"5..", %s}[$__rate_interval]
    )
  ),
  "operation",
  " ",
  "http_request_method",
  "http_route"
) or 0 * %s) / %s`, selectors, reqsRateOpsQuery, reqsRateOpsQuery)
	reqsRateOpsErrorQueryId = "reqsRateOpsError"

	p95LatencyOpsQuery = fmt.Sprintf(`
label_join(
  histogram_quantile(
    0.95,
    sum by (le, http_request_method, http_route) (rate(http_server_request_duration_seconds_bucket{%s}[5m]))
  ),
  "operation",
  " ",
  "http_request_method",
  "http_route"
)`, selectors)
	p95LatencyOpsQueryId = "p95LatencyOps"
)
