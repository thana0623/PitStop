package process

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthChecker polls health check URLs for services.
type HealthChecker struct {
	client *http.Client
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// HealthResult represents the result of a health check.
type HealthResult struct {
	Service string
	URL     string
	Ready   bool
	Err     error
}

// CheckAll polls all health check URLs concurrently.
// It returns results as each check completes.
func (hc *HealthChecker) CheckAll(ctx context.Context, services map[string]string) []HealthResult {
	var (
		mu      sync.Mutex
		results []HealthResult
		wg      sync.WaitGroup
	)

	for name, url := range services {
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			result := hc.waitForReady(ctx, name, url)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(name, url)
	}

	wg.Wait()
	return results
}

// waitForReady polls a health check URL until it returns 200 or context is cancelled.
func (hc *HealthChecker) waitForReady(ctx context.Context, name, url string) HealthResult {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// First check immediately.
	if hc.check(ctx, url) {
		return HealthResult{Service: name, URL: url, Ready: true}
	}

	for {
		select {
		case <-ctx.Done():
			return HealthResult{
				Service: name,
				URL:     url,
				Ready:   false,
				Err:     fmt.Errorf("health check timed out"),
			}
		case <-ticker.C:
			if hc.check(ctx, url) {
				return HealthResult{Service: name, URL: url, Ready: true}
			}
		}
	}
}

// check performs a single HTTP GET health check.
func (hc *HealthChecker) check(ctx context.Context, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
