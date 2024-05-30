package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListRoutes prints all registered routes in the Gin application, grouped by the first segment of the route
func ListRoutes(router *gin.Engine) {
	routeMap := make(map[string][]string)

	for _, route := range router.Routes() {
		segments := strings.Split(route.Path, "/")
		if len(segments) > 1 {
			firstSegment := segments[1]
			key := "/" + firstSegment
			routeMap[key] = append(routeMap[key], fmt.Sprintf("%s -> %s", route.Path, route.Handler))
		}
	}

	firstSegments := make([]string, 0, len(routeMap))
	for k := range routeMap {
		firstSegments = append(firstSegments, k)
	}
	sort.Strings(firstSegments)

	for _, fs := range firstSegments {
		fmt.Printf("%s:\n", fs)
		for _, route := range routeMap[fs] {
			fmt.Println("  " + route)
		}
		fmt.Println()
	}
}
