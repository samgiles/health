# Async Healthcheck controller for Go

[![Build Status](https://travis-ci.com/samgiles/health.svg?branch=master)](https://travis-ci.com/samgiles/health)

- Supports Readiness and Liveness states (k8s).
- Async healthchecks only
- Go mod support

## Example

```go
package main

import (
    "fmt"
    "time"

    "github.com/samgiles/health"
)

type HealthCheckExample struct {
    health.DefaultHealthCheck

    ready bool
    healthy bool
}

func (hc *HealthCheckExample) RunHealthCheck() health.HealthCheckResult {
    if !hc.healthy {
        return health.UnhealthyResult("Not Healthy!")
    }

    if !hc.ready {
        return health.NotReadyResult("Not Ready!")
    }

    return health.HealthyResult()
}

func (hc *HealthCheckExample) HealthCheckName() string {
	return "my-test-healthcheck"
}

func (hc *HealthCheckExample) HealthCheckFrequency() time.Duration {
	return 5 * time.Second
}

func main() {
   controller := health.NewHealthCheckController()
   defer controller.Stop()

   controller.AddHealthCheck(&HealthCheckExample{})

   // You might use the values from Readiness and Liveness in your own router
   // and response
   for {
       time.Sleep(5 * time.Second)
       fmt.Println(controller.Readiness())
       fmt.Println(controller.Liveness())
   }
}
```

# License

MIT
