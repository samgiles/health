package health

import (
	"testing"
	"time"
)

type TestHealthCheck struct {
	result         HealthCheckResult
	executedSignal chan bool
	name           string
}

func (hc *TestHealthCheck) HealthCheckComplete() {
	hc.executedSignal <- true
}

func (hc *TestHealthCheck) RunHealthCheck() HealthCheckResult {
	return hc.result
}
func (hc *TestHealthCheck) HealthCheckName() string {
	return hc.name
}

func (hc *TestHealthCheck) HealthCheckFrequency() time.Duration {
	return 100 * time.Millisecond
}

func (hc *TestHealthCheck) ShutdownHealthCheck() {}

func (hc *TestHealthCheck) InitialHealthCheckState() HealthCheckResult {
	return HealthyResult()
}

func TestHealthCheckControllerHealthy(t *testing.T) {
	testHealthCheck := TestHealthCheck{HealthyResult(), make(chan bool, 1), "test"}
	controller := NewHealthCheckController()
	defer controller.Stop()

	controller.AddHealthCheck(&testHealthCheck)

	<-testHealthCheck.executedSignal

	if controller.Readiness().Ok != true {
		t.Errorf("Expected readiness result to be ok")
	}

	if controller.Liveness().Ok != true {
		t.Error("Expected liveness result to be ok")
	}
}

func TestHealtCheckControllerNotReady(t *testing.T) {
	testHealthCheck := TestHealthCheck{NotReadyResult("not ready"), make(chan bool, 1), "test"}
	controller := NewHealthCheckController()
	defer controller.Stop()

	controller.AddHealthCheck(&testHealthCheck)

	<-testHealthCheck.executedSignal

	if controller.Readiness().Ok != false {
		t.Errorf("Expected readiness result to be not ok")
	}

	if controller.Liveness().Ok != true {
		t.Error("Expected liveness result to be ok")
	}
}

func TestHealthcheckControllerUnhealthy(t *testing.T) {
	testHealthCheck := TestHealthCheck{UnhealthyResult("not healthy"), make(chan bool, 1), "test"}
	controller := NewHealthCheckController()
	defer controller.Stop()

	controller.AddHealthCheck(&testHealthCheck)

	<-testHealthCheck.executedSignal
	if controller.Readiness().Ok != false {
		t.Errorf("Expected readiness result to be not ok")
	}

	if controller.Liveness().Ok != false {
		t.Error("Expected liveness result to be not ok")
	}
}

func TestHealthCheckManyHealthy(t *testing.T) {
	testHealthCheckA := TestHealthCheck{HealthyResult(), make(chan bool, 1), "testA"}
	testHealthCheckB := TestHealthCheck{HealthyResult(), make(chan bool, 1), "testB"}
	testHealthCheckC := TestHealthCheck{HealthyResult(), make(chan bool, 1), "testC"}

	controller := NewHealthCheckController()
	defer controller.Stop()

	controller.AddHealthCheck(&testHealthCheckA)
	controller.AddHealthCheck(&testHealthCheckB)
	controller.AddHealthCheck(&testHealthCheckC)

	// Wait for the async healthchecks to complete
	<-testHealthCheckA.executedSignal
	<-testHealthCheckB.executedSignal
	<-testHealthCheckC.executedSignal

	if controller.Readiness().Ok != true {
		t.Errorf("Expected readiness result to be ok")
	}

	if controller.Liveness().Ok != true {
		t.Error("Expected liveness result to be ok")
	}
}

func TestHealthCheckManyOneUnhealthy(t *testing.T) {
	testHealthCheckA := TestHealthCheck{HealthyResult(), make(chan bool, 1), "testA"}
	testHealthCheckB := TestHealthCheck{HealthyResult(), make(chan bool, 1), "testB"}
	testHealthCheckC := TestHealthCheck{UnhealthyResult("unhealthy"), make(chan bool, 1), "testC"}

	controller := NewHealthCheckController()
	defer controller.Stop()

	controller.AddHealthCheck(&testHealthCheckA)
	controller.AddHealthCheck(&testHealthCheckB)
	controller.AddHealthCheck(&testHealthCheckC)

	// Wait for the async healthchecks to complete
	<-testHealthCheckA.executedSignal
	<-testHealthCheckB.executedSignal
	<-testHealthCheckC.executedSignal

	if controller.Readiness().Ok != false {
		t.Errorf("Expected readiness result to be not ok")
	}

	if controller.Liveness().Ok != false {
		t.Error("Expected liveness result to be not ok")
	}
}
