package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
	"github.com/Rowlyge/kuflow/internal/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentHealthEndpoint(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	resp, err := env.Client().Get(
		env.URL() + "/health",
	)
	if err != nil {
		t.Fatalf("GET /health error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"GET /health status = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	if requestID := resp.Header.Get("X-Request-ID"); requestID == "" {
		t.Fatal("GET /health did not return X-Request-ID")
	}

	var body map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf(
			"decode /health response: %v",
			err,
		)
	}

	if body["status"] != "ok" {
		t.Fatalf(
			"GET /health response status = %v, want %q",
			body["status"],
			"ok",
		)
	}
}

func TestEnvironmentProxyPipeline(t *testing.T) {
	var requests atomic.Int64

	upstream := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requests.Add(1)

			if got := r.Header.Get("X-Forwarded-Proto"); got != "http" {
				t.Errorf(
					"X-Forwarded-Proto = %q, want %q",
					got,
					"http",
				)
			}

			if got := r.Header.Get("X-Forwarded-Host"); got == "" {
				t.Error("X-Forwarded-Host is empty")
			}

			w.Header().Set("X-Upstream", "integration")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("upstream response"))
		}),
	)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test?value=42",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-valid-key",
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(
			"proxy response status = %d, want %d",
			resp.StatusCode,
			http.StatusCreated,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read proxy response: %v", err)
	}

	if string(body) != "upstream response" {
		t.Fatalf(
			"proxy response body = %q, want %q",
			string(body),
			"upstream response",
		)
	}

	if got := resp.Header.Get("X-KuFlow"); got != "true" {
		t.Fatalf(
			"X-KuFlow = %q, want %q",
			got,
			"true",
		)
	}

	if requestID := resp.Header.Get("X-Request-ID"); requestID == "" {
		t.Fatalf("proxy response did not contain X-Request-ID")
	}

	if got := resp.Header.Get("X-Upstream"); got != "integration" {
		t.Fatalf(
			"X-Upstream = %q, want %q",
			got,
			"integration",
		)
	}

	if got := requests.Load(); got != 1 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			1,
		)
	}
}

func newAuthTestUpstream(
	t *testing.T,
	requests *atomic.Int64,
) *httptest.Server {
	t.Helper()

	return httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requests.Add(1)

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("upstream response"))
		}),
	)
}

func newValidAPIKey() authcache.APIKey {
	return authcache.APIKey{
		ID:      1,
		Key:     "integration-valid-key",
		Owner:   "integration-test",
		Enabled: true,
	}
}

func TestEnvironmentAuthenticationValidKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-valid-key",
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusCreated,
		)
	}

	if got := requests.Load(); got != 1 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			1,
		)
	}
}

func TestEnvironmentAuthenticationMissingKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	resp, err := env.Client().Get(
		env.URL() + "/proxy-test",
	)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}

func TestEnvironmentAuthenticationInvalidKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-invalid-key",
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}

func TestEnvironmentAuthenticationDisabledKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	key := newValidAPIKey()
	key.Enabled = false

	env.SetAPIKeys(key)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		key.Key,
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}

func TestEnvironmentAuthenticationExpiredKey(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	expiredAt := time.Now().Add(-time.Minute)

	key := newValidAPIKey()
	key.ExpiresAt = &expiredAt

	env.SetAPIKeys(key)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		key.Key,
	)

	resp, err := env.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			0,
		)
	}
}

func TestEnvironmentRateLimitPipeline(t *testing.T) {
	var requests atomic.Int64

	upstream := newAuthTestUpstream(t, &requests)
	defer upstream.Close()

	env, err := NewEnvironmentWithLimits(
		ratelimit.Config{
			Capacity:       2,
			RefillTokens:   2,
			RefillInterval: time.Hour,
		},
		100,
		upstream.URL,
	)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	client := env.Client()

	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(
			http.MethodGet,
			env.URL()+"/proxy-test",
			nil,
		)
		if err != nil {
			t.Fatalf("create request %d: %v", i+1, err)
		}

		req.Header.Set(
			"X-API-Key",
			"integration-valid-key",
		)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf(
				"request %d error = %v",
				i+1,
				err,
			)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf(
				"request %d status = %d, want %d",
				i+1,
				resp.StatusCode,
				http.StatusCreated,
			)
		}

		resp.Body.Close()
	}

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create rejected request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-valid-key",
	)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("rejected request error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf(
			"rejected request status = %d, want %d",
			resp.StatusCode,
			http.StatusTooManyRequests,
		)
	}

	if got := requests.Load(); got != 2 {
		t.Fatalf(
			"upstream request count = %d, want %d",
			got,
			2,
		)
	}
}

func TestEnvironmentConnectionLimitPipeline(t *testing.T) {
	var requests atomic.Int64

	started := make(chan struct{}, 2)
	release := make(chan struct{})

	upstream := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requests.Add(1)

			started <- struct{}{}

			<-release

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("upstream response"))
		}),
	)
	defer upstream.Close()

	env, err := NewEnvironmentWithLimits(
		ratelimit.Config{
			Capacity:       100,
			RefillTokens:   100,
			RefillInterval: time.Minute,
		},
		2,
		upstream.URL,
	)
	if err != nil {
		t.Fatalf("NewEnvironmentWithConnectionLimit() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	var releaseOnce sync.Once
	defer releaseOnce.Do(func() {
		close(release)
	})

	type result struct {
		status int
		err    error
	}

	results := make(chan result, 2)

	sendRequest := func() {
		req, err := http.NewRequest(
			http.MethodGet,
			env.URL()+"/proxy-test",
			nil,
		)
		if err != nil {
			results <- result{err: err}
			return
		}

		req.Header.Set(
			"X-API-Key",
			"integration-valid-key",
		)

		resp, err := env.Client().Do(req)
		if err != nil {
			results <- result{err: err}
			return
		}

		defer resp.Body.Close()

		results <- result{
			status: resp.StatusCode,
		}
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		sendRequest()
	}()

	go func() {
		defer wg.Done()
		sendRequest()
	}()

	// Wait until both connection-limit slots are occupied
	// by requests that have actually reached the upstream.
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-timeout.C:
			t.Fatal("timed out waiting for both upstream requests")
		}
	}

	// Both connection slots are now occupied.
	// The third request must be rejected before reaching the upstream.
	thirdReq, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create third request: %v", err)
	}

	thirdReq.Header.Set(
		"X-API-Key",
		"integration-valid-key",
	)

	thirdResp, err := env.Client().Do(thirdReq)
	if err != nil {
		t.Fatalf("third request error = %v", err)
	}
	defer thirdResp.Body.Close()

	if thirdResp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf(
			"third request status = %d, want %d",
			thirdResp.StatusCode,
			http.StatusTooManyRequests,
		)
	}

	// The third request must not reach the upstream.
	if got := requests.Load(); got != 2 {
		t.Fatalf(
			"upstream request count after rejection = %d, want %d",
			got,
			2,
		)
	}

	// Release the first two upstream requests.
	releaseOnce.Do(func() {
		close(release)
	})

	wg.Wait()

	for i := 0; i < 2; i++ {
		select {
		case result := <-results:
			if result.err != nil {
				t.Fatalf(
					"upstream request %d error = %v",
					i+1,
					result.err,
				)
			}

			if result.status != http.StatusCreated {
				t.Fatalf(
					"upstream request %d status = %d, want %d",
					i+1,
					result.status,
					http.StatusCreated,
				)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf(
				"timed out waiting for upstream request %d",
				i+1,
			)
		}
	}

	if got := requests.Load(); got != 2 {
		t.Fatalf(
			"final upstream request count = %d, want %d",
			got,
			2,
		)
	}
}

func TestEnvironmentBalancerRoundRobin(t *testing.T) {
	var requests [3]atomic.Int64

	upstreams := make([]*httptest.Server, 0, 3)

	for i := range requests {
		index := i

		upstream := httptest.NewServer(
			http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				requests[index].Add(1)

				w.Header().Set(
					"X-Upstream",
					fmt.Sprintf("upstream-%d", index+1),
				)

				w.WriteHeader(http.StatusCreated)

				_, _ = w.Write(
					[]byte(
						fmt.Sprintf(
							"response-from-upstream-%d",
							index+1,
						),
					),
				)
			}),
		)

		upstreams = append(upstreams, upstream)
	}

	defer func() {
		for _, upstream := range upstreams {
			upstream.Close()
		}
	}()

	targets := make([]string, 0, len(upstreams))

	for _, upstream := range upstreams {
		targets = append(targets, upstream.URL)
	}

	env, err := NewEnvironment(targets...)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	expectedUpstreams := []string{
		"upstream-1",
		"upstream-2",
		"upstream-3",
		"upstream-1",
		"upstream-2",
		"upstream-3",
	}

	for i, expected := range expectedUpstreams {
		req, err := http.NewRequest(
			http.MethodGet,
			env.URL()+"/proxy-test",
			nil,
		)
		if err != nil {
			t.Fatalf(
				"request %d: create request: %v",
				i+1,
				err,
			)
		}

		req.Header.Set(
			"X-API-Key",
			"integration-valid-key",
		)

		resp, err := env.Client().Do(req)
		if err != nil {
			t.Fatalf(
				"request %d: proxy request error: %v",
				i+1,
				err,
			)
		}

		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()

			t.Fatalf(
				"request %d: status = %d, want %d",
				i+1,
				resp.StatusCode,
				http.StatusCreated,
			)
		}

		if got := resp.Header.Get("X-Upstream"); got != expected {
			resp.Body.Close()

			t.Fatalf(
				"request %d: X-Upstream = %q, want %q",
				i+1,
				got,
				expected,
			)
		}

		resp.Body.Close()
	}

	for i := range requests {
		if got := requests[i].Load(); got != 2 {
			t.Fatalf(
				"upstream-%d request count = %d, want %d",
				i+1,
				got,
				2,
			)
		}
	}
}

func TestEnvironmentHealthCheckerRemovesUnhealthyUpstream(
	t *testing.T,
) {
	var upstreamAHealthy atomic.Bool
	upstreamAHealthy.Store(true)

	upstreamA := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			switch r.URL.Path {

			case "/health":
				if !upstreamAHealthy.Load() {
					http.Error(
						w,
						"unhealthy",
						http.StatusServiceUnavailable,
					)
					return
				}

				w.WriteHeader(http.StatusOK)

			default:
				w.Header().Set(
					"X-Upstream",
					"upstream-A",
				)
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte("response-from-upstream-A"),
				)
			}
		}),
	)
	defer upstreamA.Close()

	upstreamB := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			switch r.URL.Path {

			case "/health":
				w.WriteHeader(http.StatusOK)

			default:
				w.Header().Set(
					"X-Upstream",
					"upstream-B",
				)
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte("response-from-upstream-B"),
				)
			}
		}),
	)
	defer upstreamB.Close()

	env, err := NewEnvironmentWithHealth(
		upstreamA.URL,
		upstreamB.URL,
	)
	if err != nil {
		t.Fatalf(
			"NewEnvironmentWithHealth() error = %v",
			err,
		)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	// Wait until the initial health checks confirm
	// that both upstreams are healthy.
	deadline := time.Now().Add(2 * time.Second)

	for {
		upstreams := env.upstreams.Upstreams()

		if len(upstreams) != 2 {
			t.Fatalf(
				"upstream count = %d, want %d",
				len(upstreams),
				2,
			)
		}

		if upstreams[0].Alive() &&
			upstreams[1].Alive() {
			break
		}

		if time.Now().After(deadline) {
			t.Fatal(
				"timed out waiting for both upstreams to become healthy",
			)
		}

		time.Sleep(5 * time.Millisecond)
	}

	// Both upstreams are healthy, so round-robin must be
	// able to send traffic to both of them.
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(
			http.MethodGet,
			env.URL()+"/proxy-test",
			nil,
		)
		if err != nil {
			t.Fatalf(
				"create request %d: %v",
				i+1,
				err,
			)
		}

		req.Header.Set(
			"X-API-Key",
			"integration-valid-key",
		)

		resp, err := env.Client().Do(req)
		if err != nil {
			t.Fatalf(
				"request %d error = %v",
				i+1,
				err,
			)
		}

		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()

			t.Fatalf(
				"request %d status = %d, want %d",
				i+1,
				resp.StatusCode,
				http.StatusCreated,
			)
		}

		resp.Body.Close()
	}

	// Make upstream A unhealthy.
	upstreamAHealthy.Store(false)

	// Wait until Health Checker detects the failure
	// and marks upstream A as DOWN.
	deadline = time.Now().Add(2 * time.Second)

	for env.upstreams.Upstreams()[0].Alive() {
		if time.Now().After(deadline) {
			t.Fatal(
				"timed out waiting for upstream-A to become unhealthy",
			)
		}

		time.Sleep(5 * time.Millisecond)
	}

	if !env.upstreams.Upstreams()[1].Alive() {
		t.Fatal("upstream-B became unhealthy unexpectedly")
	}

	// After A is marked DOWN, all traffic must go to B.
	for i := 0; i < 6; i++ {
		req, err := http.NewRequest(
			http.MethodGet,
			env.URL()+"/proxy-test",
			nil,
		)
		if err != nil {
			t.Fatalf(
				"create request %d: %v",
				i+1,
				err,
			)
		}

		req.Header.Set(
			"X-API-Key",
			"integration-valid-key",
		)

		resp, err := env.Client().Do(req)
		if err != nil {
			t.Fatalf(
				"request %d error = %v",
				i+1,
				err,
			)
		}

		if resp.StatusCode != http.StatusCreated {
			resp.Body.Close()

			t.Fatalf(
				"request %d status = %d, want %d",
				i+1,
				resp.StatusCode,
				http.StatusCreated,
			)
		}

		if got := resp.Header.Get("X-Upstream"); got != "upstream-B" {
			resp.Body.Close()

			t.Fatalf(
				"request %d X-Upstream = %q, want %q",
				i+1,
				got,
				"upstream-B",
			)
		}

		resp.Body.Close()
	}
}

func TestEnvironmentCircuitBreakerStopsUpstreamTraffic(t *testing.T) {
	var requests atomic.Int64

	upstream := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requests.Add(1)

			http.Error(
				w,
				"upstream failure",
				http.StatusInternalServerError,
			)
		}),
	)
	defer upstream.Close()

	env, err := NewEnvironment(upstream.URL)
	if err != nil {
		t.Fatalf("NewEnvironment() error = %v", err)
	}
	defer env.Close()

	env.SetAPIKeys(
		newValidAPIKey(),
	)

	client := env.Client()

	// Default Circuit Breaker configuration:
	// FailureThreshold = 5.
	const failureThreshold = 5

	for i := 0; i < failureThreshold; i++ {
		req, err := http.NewRequest(
			http.MethodGet,
			env.URL()+"/proxy-test",
			nil,
		)
		if err != nil {
			t.Fatalf(
				"request %d: create request: %v",
				i+1,
				err,
			)
		}

		req.Header.Set(
			"X-API-Key",
			"integration-valid-key",
		)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf(
				"request %d: proxy request error: %v",
				i+1,
				err,
			)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			resp.Body.Close()

			t.Fatalf(
				"request %d: status = %d, want %d",
				i+1,
				resp.StatusCode,
				http.StatusInternalServerError,
			)
		}

		resp.Body.Close()
	}

	// All failures must have reached the upstream.
	if got := requests.Load(); got != failureThreshold {
		t.Fatalf(
			"upstream request count after failures = %d, want %d",
			got,
			failureThreshold,
		)
	}

	// The next request must be rejected by the opened
	// Circuit Breaker without reaching the upstream.
	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	if err != nil {
		t.Fatalf("create rejected request: %v", err)
	}

	req.Header.Set(
		"X-API-Key",
		"integration-valid-key",
	)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("rejected request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"rejected request status = %d, want %d",
			resp.StatusCode,
			http.StatusServiceUnavailable,
		)
	}

	// Crucial assertion:
	// the rejected request must NOT reach the upstream.
	if got := requests.Load(); got != failureThreshold {
		t.Fatalf(
			"upstream request count after breaker opened = %d, want %d",
			got,
			failureThreshold,
		)
	}
}

func TestEnvironmentTransportConnectionFailure(t *testing.T) {
	t.Parallel()

	env, err := NewEnvironment(
		"http://127.0.0.1:65534",
	)
	require.NoError(t, err)
	defer env.Close()

	expiresAt := time.Now().Add(time.Hour)

	env.SetAPIKeys(
		authcache.APIKey{
			Key:       "test-key",
			Owner:     "integration-test",
			Enabled:   true,
			ExpiresAt: &expiresAt,
		},
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("X-API-Key", "test-key")

	resp, err := env.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(
		t,
		http.StatusBadGateway,
		resp.StatusCode,
	)
}

func TestEnvironmentTransportTimeout(t *testing.T) {
	t.Parallel()

	slowUpstream := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)

			w.WriteHeader(http.StatusOK)
		}),
	)
	defer slowUpstream.Close()

	env, err := NewEnvironment(
		slowUpstream.URL,
	)
	require.NoError(t, err)
	defer env.Close()

	expiresAt := time.Now().Add(time.Hour)

	env.SetAPIKeys(
		authcache.APIKey{
			Key:       "test-key",
			Owner:     "integration-test",
			Enabled:   true,
			ExpiresAt: &expiresAt,
		},
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("X-API-Key", "test-key")

	resp, err := env.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(
		t,
		http.StatusBadGateway,
		resp.StatusCode,
	)
}

func TestEnvironmentTransportPropagatesResponse(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Upstream", "ok")
			w.WriteHeader(http.StatusCreated)

			_, _ = w.Write([]byte("transport-ok"))
		}),
	)
	defer upstream.Close()

	env, err := NewEnvironment(
		upstream.URL,
	)
	require.NoError(t, err)
	defer env.Close()

	expiresAt := time.Now().Add(time.Hour)

	env.SetAPIKeys(
		authcache.APIKey{
			Key:       "test-key",
			Owner:     "integration-test",
			Enabled:   true,
			ExpiresAt: &expiresAt,
		},
	)

	req, err := http.NewRequest(
		http.MethodGet,
		env.URL()+"/proxy-test",
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("X-API-Key", "test-key")

	resp, err := env.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(
		t,
		http.StatusCreated,
		resp.StatusCode,
	)

	assert.Equal(
		t,
		"ok",
		resp.Header.Get("X-Upstream"),
	)

	assert.Equal(
		t,
		"transport-ok",
		string(body),
	)
}
