package handler

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"runtime"
)

// registerDebugRoutes registers /debug/* endpoints (pprof, expvar, and
// custom diagnostics). Gated on cfg.Server.Debug, which already encodes the
// full --debug/DEBUG/MODE=debug precedence chain (AI.md PART 6). When debug
// mode is off, none of these patterns are registered and the mux returns
// its normal 404 for any /debug/* request.
func (rt *Router) registerDebugRoutes() {
	if !rt.cfg.Server.Debug {
		return
	}

	// pprof endpoints
	rt.hf("GET /debug/pprof/", pprof.Index)
	rt.hf("GET /debug/pprof/cmdline", pprof.Cmdline)
	rt.hf("GET /debug/pprof/profile", pprof.Profile)
	rt.hf("GET /debug/pprof/symbol", pprof.Symbol)
	rt.hf("GET /debug/pprof/trace", pprof.Trace)
	rt.h("GET /debug/pprof/heap", pprof.Handler("heap"))
	rt.h("GET /debug/pprof/goroutine", pprof.Handler("goroutine"))
	rt.h("GET /debug/pprof/allocs", pprof.Handler("allocs"))
	rt.h("GET /debug/pprof/block", pprof.Handler("block"))
	rt.h("GET /debug/pprof/mutex", pprof.Handler("mutex"))
	rt.h("GET /debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	// expvar
	rt.h("GET /debug/vars", expvar.Handler())

	// Custom debug endpoints
	rt.hf("GET /debug/config", rt.handleDebugConfig)
	rt.hf("GET /debug/routes", rt.handleDebugRoutes)
	rt.hf("GET /debug/cache", rt.handleDebugCache)
	rt.hf("GET /debug/db", rt.handleDebugDB)
	rt.hf("GET /debug/scheduler", rt.handleDebugScheduler)
	rt.hf("GET /debug/memory", rt.handleDebugMemory)
	rt.hf("GET /debug/goroutines", rt.handleDebugGoroutines)
}

// handleDebugConfig returns the current configuration with sensitive values
// redacted (AI.md PART 6, Debug API Endpoints).
func (rt *Router) handleDebugConfig(w http.ResponseWriter, r *http.Request) {
	cfg := rt.cfg
	respondJSON(w, http.StatusOK, map[string]any{
		"server": map[string]any{
			"address":     cfg.Server.Address,
			"port":        cfg.Server.Port,
			"fqdn":        cfg.Server.FQDN,
			"mode":        cfg.Server.Mode,
			"debug":       cfg.Server.Debug,
			"admin_path":  cfg.Server.AdminPath,
			"api_version": cfg.Server.APIVersion,
		},
		"database": map[string]any{
			"driver": cfg.Database.Driver,
			"host":   cfg.Database.Host,
			"port":   cfg.Database.Port,
			"name":   cfg.Database.Name,
			"user":   cfg.Database.User,
			// password redacted
		},
		"email": map[string]any{
			"enabled":   cfg.Email.Enabled,
			"host":      cfg.Email.Host,
			"port":      cfg.Email.Port,
			"username":  cfg.Email.Username,
			"from":      cfg.Email.From,
			"from_name": cfg.Email.FromName,
			"tls":       cfg.Email.TLS,
			"tls_mode":  cfg.Email.TLSMode,
			// password redacted
		},
		"logging": map[string]any{
			"level":  cfg.Logging.Level,
			"format": cfg.Logging.Format,
		},
	})
}

// handleDebugRoutes returns every route pattern registered on this router
// (AI.md PART 6, Debug API Endpoints).
func (rt *Router) handleDebugRoutes(w http.ResponseWriter, r *http.Request) {
	routes := make([]map[string]string, 0, len(rt.routes))
	for _, pattern := range rt.routes {
		method, path := pattern, pattern
		if i := indexOfSpace(pattern); i >= 0 {
			method, path = pattern[:i], pattern[i+1:]
		} else {
			method = ""
		}
		routes = append(routes, map[string]string{
			"method": method,
			"route":  path,
		})
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"count":  len(routes),
		"routes": routes,
	})
}

// indexOfSpace returns the index of the first space in s, or -1 if none.
func indexOfSpace(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			return i
		}
	}
	return -1
}

// handleDebugCache returns cache statistics. This project has no in-process
// cache layer yet, so this reports that explicitly rather than fabricating
// values (AI.md PART 6, Debug API Endpoints).
func (rt *Router) handleDebugCache(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"available": false,
		"message":   "no in-process cache layer is configured",
	})
}

// handleDebugDB returns database connection pool statistics (AI.md PART 6,
// Debug API Endpoints).
func (rt *Router) handleDebugDB(w http.ResponseWriter, r *http.Request) {
	if rt.db == nil {
		respondJSON(w, http.StatusOK, map[string]any{"available": false})
		return
	}

	stats := rt.db.Stats()
	respondJSON(w, http.StatusOK, map[string]any{
		"open_connections":    stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration_ms":    stats.WaitDuration.Milliseconds(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	})
}

// handleDebugScheduler returns background task status. The scheduler is not
// currently instantiated in the running process, so this reports that
// explicitly rather than fabricating task state (AI.md PART 6, Debug API
// Endpoints).
func (rt *Router) handleDebugScheduler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"available": false,
		"message":   "no scheduler is running in this process",
	})
}

// handleDebugMemory returns runtime memory statistics (AI.md PART 6, Debug
// Implementation).
func (rt *Router) handleDebugMemory(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	respondJSON(w, http.StatusOK, map[string]any{
		"alloc_mb":       m.Alloc / 1024 / 1024,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"num_gc":         m.NumGC,
		"heap_objects":   m.HeapObjects,
		"goroutines":     runtime.NumGoroutine(),
	})
}

// handleDebugGoroutines returns goroutine stack traces (AI.md PART 6, Debug
// Implementation).
func (rt *Router) handleDebugGoroutines(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, true)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(buf[:n])
}
