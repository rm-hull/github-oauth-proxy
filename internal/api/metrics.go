package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	callersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "github_oauth_proxy_callers_total",
			Help: "Total number of requests by caller IP address",
		},
		[]string{"ip", "error"},
	)

	methodCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "github_oauth_proxy_method_calls_total",
			Help: "Total number of method calls by client_id and method",
		},
		[]string{"client_id", "method", "error"},
	)
)

func callersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		c.Next()
		errorLabel := fmt.Sprintf("%t", c.Writer.Status() >= http.StatusBadRequest)
		callersTotal.WithLabelValues(ip, errorLabel).Inc()
	}
}

func incrementMethodCall(clientID, method string, isError bool) {
	errorLabel := fmt.Sprintf("%t", isError)
	methodCallsTotal.WithLabelValues(clientID, method, errorLabel).Inc()
}
