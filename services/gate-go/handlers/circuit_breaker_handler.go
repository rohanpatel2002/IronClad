package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// CircuitBreakerStatus is a snapshot of a circuit breaker's state.
type CircuitBreakerStatus struct {
	Name               string    `json:"name"`
	State              string    `json:"state"`
	Requests           uint32    `json:"requests_total"`
	TotalSuccesses     uint32    `json:"successes"`
	TotalFailures      uint32    `json:"failures"`
	ConsecutiveFailures uint32   `json:"consecutive_failures"`
	LastStateChange    time.Time `json:"last_state_change"`
}

// circuitBreakerRegistry holds references to all registered circuit breakers.
var circuitBreakerRegistry = map[string]*gobreaker.CircuitBreaker{}

// RegisterCircuitBreaker adds a circuit breaker to the status registry.
func RegisterCircuitBreaker(name string, cb *gobreaker.CircuitBreaker) {
	circuitBreakerRegistry[name] = cb
}

// CircuitBreakerStatusHandler returns the HTTP handler for CB status.
func CircuitBreakerStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		statuses := make([]CircuitBreakerStatus, 0, len(circuitBreakerRegistry))

		for name, cb := range circuitBreakerRegistry {
			counts := cb.Counts()
			state := "closed"
			switch cb.State() {
			case gobreaker.StateOpen:
				state = "open"
			case gobreaker.StateHalfOpen:
				state = "half-open"
			}
			statuses = append(statuses, CircuitBreakerStatus{
				Name:                name,
				State:               state,
				Requests:            counts.Requests,
				TotalSuccesses:      counts.TotalSuccesses,
				TotalFailures:       counts.TotalFailures,
				ConsecutiveFailures: counts.ConsecutiveFailures,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"circuit_breakers": statuses,
			"total":            len(statuses),
			"timestamp":        time.Now().UTC(),
		})
	}
}
