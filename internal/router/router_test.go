package router

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitRouter_RegistersCoreRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	InitRouter(r)

	type wantRoute struct {
		method string
		path   string
	}
	wants := []wantRoute{
		{method: "GET", path: "/api/ping"},
		{method: "POST", path: "/api/login"},
		{method: "POST", path: "/api/register"},
		{method: "GET", path: "/api/user/ping"},
		{method: "GET", path: "/api/admin/stats"},
	}

	have := make(map[string]bool)
	for _, rt := range r.Routes() {
		have[rt.Method+" "+rt.Path] = true
	}

	for _, w := range wants {
		if !have[w.method+" "+w.path] {
			t.Fatalf("missing route: %s %s", w.method, w.path)
		}
	}
}
