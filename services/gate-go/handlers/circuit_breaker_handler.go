package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
var (
	circuitBreakerRegistry = map[string]*gobreaker.CircuitBreaker{}
	
	cbStateMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ironclad_gate_circuit_breaker_state",
			Help: "Current state of the circuit breaker (0=closed, 1=half-open, 2=open)",
		},
		[]string{"name"},
	)
)

// RegisterCircuitBreaker adds a circuit breaker to the status registry.
func RegisterCircuitBreaker(name string, cb *gobreaker.CircuitBreaker) {
	circuitBreakerRegistry[name] = cb
}

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			for name, cb := range circuitBreakerRegistry {
				state := 0.0
				switch cb.State() {
				case gobreaker.StateHalfOpen:
					state = 1.0
				case gobreaker.StateOpen:
					state = 2.0
				}
				cbStateMetric.WithLabelValues(name).Set(state)
			}
		}
	}()
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

