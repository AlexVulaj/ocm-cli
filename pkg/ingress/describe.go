package ingress

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/openshift-online/ocm-cli/pkg/utils"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func PrintIngressDescription(ingress *cmv1.Ingress, cluster *cmv1.Cluster) error {
	entries := generateEntriesOutput(cluster, ingress)
	ingressOutput := ""
	keys := utils.MapKeys(entries)
	sort.Strings(keys)
	minWidth := getMinWidth(keys)
	for _, key := range keys {
		ingressOutput += fmt.Sprintf("%s: %s\n", key, strings.Repeat(" ", minWidth-len(key))+entries[key])
	}
	fmt.Print(ingressOutput)
	return nil
}

// Min width is defined as the length of the longest string
func getMinWidth(keys []string) int {
	minWidth := 0
	for _, key := range keys {
		if len(key) > minWidth {
			minWidth = len(key)
		}
	}
	return minWidth
}

func generateEntriesOutput(cluster *cmv1.Cluster, ingress *cmv1.Ingress) map[string]string {
	private := false
	if ingress.Listening() == cmv1.ListeningMethodInternal {
		private = true
	}
	entries := map[string]string{
		"ID":         ingress.ID(),
		"Cluster ID": cluster.ID(),
		"Default":    strconv.FormatBool(ingress.Default()),
		"Private":    strconv.FormatBool(private),
		"LB-Type":    string(ingress.LoadBalancerType()),
	}
	// These are only available for ingress v2
	wildcardPolicy := string(ingress.RouteWildcardPolicy())
	if wildcardPolicy != "" {
		entries["Wildcard Policy"] = string(ingress.RouteWildcardPolicy())
	}
	namespaceOwnershipPolicy := string(ingress.RouteNamespaceOwnershipPolicy())
	if namespaceOwnershipPolicy != "" {
		entries["Namespace Ownership Policy"] = namespaceOwnershipPolicy
	}
	routeSelectors := ""
	if len(ingress.RouteSelectors()) > 0 {
		routeSelectors = fmt.Sprintf("%v", ingress.RouteSelectors())
	}
	if routeSelectors != "" {
		entries["Route Selectors"] = routeSelectors
	}
	excludedNamespaces := utils.SliceToSortedString(ingress.ExcludedNamespaces())
	if excludedNamespaces != "" {
		entries["Excluded Namespaces"] = excludedNamespaces
	}
	componentRoutes := ""
	componentKeys := utils.MapKeys(ingress.ComponentRoutes())
	sort.Strings(componentKeys)
	for _, component := range componentKeys {
		value := ingress.ComponentRoutes()[component]
		keys := utils.MapKeys(entries)
		minWidth := getMinWidth(keys)
		depth := 4
		componentRouteEntries := map[string]string{
			"Hostname":       value.Hostname(),
			"TLS Secret Ref": value.TlsSecretRef(),
		}
		componentRoutes += fmt.Sprintf("%s: \n", strings.Repeat(" ", depth)+component)
		depth *= 2
		paramKeys := utils.MapKeys(componentRouteEntries)
		sort.Strings(paramKeys)
		for _, param := range paramKeys {
			componentRoutes += fmt.Sprintf(
				"%s: %s\n",
				strings.Repeat(" ", depth)+param,
				strings.Repeat(" ", minWidth-len(param)-depth)+componentRouteEntries[param],
			)
		}
	}
	if componentRoutes != "" {
		componentRoutes = fmt.Sprintf("\n%s", componentRoutes)
		//remove extra \n at the end
		componentRoutes = componentRoutes[:len(componentRoutes)-1]
		entries["Component Routes"] = componentRoutes
	}
	return entries
}
