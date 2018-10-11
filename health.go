package health

import (
	"sync"
	"time"
)

type HealthCheckResultType int

const (
	Healthy HealthCheckResultType   = 0x1
	Unhealthy = 0x2
	NotReady  = 0x4
)

// Indicates the result of a healthcheck along with an additional message
type HealthCheckResult struct {
	result  HealthCheckResultType
	message string
}

// Get a string representation of the HealthCheckResultType
func (r *HealthCheckResult) Result() string {
	switch r.result {
	case Healthy:
		return "healthy"
	case Unhealthy:
		return "unhealthy"
	case NotReady:
		return "not-ready"
	default:
		return "unknown"
	}
}

// Create a HealthyResult
func HealthyResult() HealthCheckResult {
	return HealthCheckResult{result: Healthy}
}

// Create an Unhealthy result
func UnhealthyResult(message string) HealthCheckResult {
	return HealthCheckResult{Unhealthy, message}
}

// Create a Not Ready result
func NotReadyResult(message string) HealthCheckResult {
	return HealthCheckResult{NotReady, message}
}

// Defines the interface for a HealthCheck.
// You can include a DefaultHealthCheck in your `struct`
// which implements NoOps for `HealthCheckComplete`, and `ShutdownHealthCheck`
// and sets the `InitialHealthCheckState` to `NotReadyResult("initial result")`
type HealthCheck interface {
	RunHealthCheck() HealthCheckResult
	HealthCheckComplete()
	HealthCheckName() string
	HealthCheckFrequency() time.Duration
	ShutdownHealthCheck()
	InitialHealthCheckState() HealthCheckResult
}

type DefaultHealthCheck struct{}

func (hc *DefaultHealthCheck) HealthCheckComplete() {}
func (hc *DefaultHealthCheck) ShutdownHealthCheck() {}
func (hc *DefaultHealthCheck) InitialHealthCheckState() HealthCheckResult {
	return NotReadyResult("initial result")
}

type HealthCheckController struct {
	checks      map[string]HealthCheck
	resultCache map[string]HealthCheckResult
	cacheLock   *sync.RWMutex
	tickers     []*time.Ticker
}

func NewHealthCheckController() HealthCheckController {
	controller := HealthCheckController{}

	controller.checks = make(map[string]HealthCheck)
	controller.resultCache = make(map[string]HealthCheckResult)
	controller.cacheLock = &sync.RWMutex{}
	return controller
}

func (hc *HealthCheckController) AddHealthCheck(healthCheck HealthCheck) {
	hc.checks[healthCheck.HealthCheckName()] = healthCheck

	hcTicker := time.NewTicker(healthCheck.HealthCheckFrequency())
	hc.tickers = append(hc.tickers, hcTicker)

	hc.resultCache[healthCheck.HealthCheckName()] = healthCheck.InitialHealthCheckState()

	go func() {
		for range hcTicker.C {
			name := healthCheck.HealthCheckName()
			result := healthCheck.RunHealthCheck()

            hc.writeResult(name, result)

			healthCheck.HealthCheckComplete()
		}
	}()
}

func (hc *HealthCheckController) writeResult(name string, result HealthCheckResult) {
    defer hc.cacheLock.Unlock()
    hc.cacheLock.Lock()

	hc.resultCache[name] = result
}

type HealthResponse struct {
	Ok          bool              `json:"ok"`
	NotOkChecks map[string]string `json:"not_ok_checks"`
}

// Get the healthcheck status for Readiness checks, NotReady or Unhealthy
// checks indicate not ok readiness
func (hc *HealthCheckController) Readiness() HealthResponse {
	return hc.buildHealthResponse(NotReady | Unhealthy)
}

// Get healthcheck status for Liveness checks (Unhealthy checks indicate not ok
// liveness)
func (hc *HealthCheckController) Liveness() HealthResponse {
	return hc.buildHealthResponse(Unhealthy)
}

func (hc *HealthCheckController) buildHealthResponse(healthType HealthCheckResultType) HealthResponse {
    defer hc.cacheLock.RUnlock()
    hc.cacheLock.RLock()

	var response = HealthResponse{true, map[string]string{}}

	for check, result := range hc.resultCache {
		if (result.result & healthType) == result.result {
			response.Ok = false
			response.NotOkChecks[check] = result.message
		}
	}

	return response
}

func (hc *HealthCheckController) Stop() {
	for _, ticker := range hc.tickers {
		ticker.Stop()
	}

	for _, check := range hc.checks {
		check.ShutdownHealthCheck()
	}
}
