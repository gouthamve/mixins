package main

import (
	"encoding/json"
	"fmt"

	"github.com/gouthamve/mixins/common"
	"github.com/gouthamve/mixins/otel-app-semantic/dashboard"
)

func main() {
	d, err := dashboard.Build(dashboard.Config{
		LogsQuery: `{service_namespace=~"$service_namespace", service_name=~"$service_name"}`,
		// ServiceNamespaces: []string{"beaverhabits"},
	})
	if err != nil {
		panic(err)
	}

	manifest := common.DashboardManifest("", d)
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(manifestJSON))
}
