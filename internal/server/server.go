package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/alert"
	"github.com/alexandrmotologa/honeyagent/internal/canary"
	"github.com/alexandrmotologa/honeyagent/internal/config"
	"github.com/alexandrmotologa/honeyagent/internal/countermeasure"
	"github.com/alexandrmotologa/honeyagent/internal/decoy"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

// Server coordinates the management API, deception studio, and proxy gateway.
type Server struct {
	cfg         *config.Config
	apiServer   *http.Server
	proxyServer *http.Server
	seeder      *canary.Seeder
	tripwire    *detector.Tripwire
}

// NewServer initializes all engine subsystems.
func NewServer(cfg *config.Config) (*Server, error) {
	// 1. Canary Engine
	tokenMgr := canary.NewTokenManager(cfg.Canary.MasterKey, cfg.Canary.TokenPrefix, cfg.Canary.TokenTTL)
	synth := canary.NewSynthesizer(tokenMgr)
	memStore := canary.NewMemoryVectorStore(cfg.VectorDBs.Memory.VectorDim)
	seeder := canary.NewSeeder(cfg, synth, memStore)

	// 2. Decoy Tool Catalog & Interceptor
	catalog := decoy.NewToolCatalog()
	interceptor := decoy.NewInterceptor(catalog)

	// 3. Detector Scanner & Tripwire State Machine
	scanner := detector.NewStreamScanner(tokenMgr, cfg.Canary.TokenPrefix)
	tripwire := detector.NewTripwire()

	// 4. Countermeasures & Deception Sandbox
	tarpit := countermeasure.NewTarpit(&cfg.Tarpit)
	sandbox := countermeasure.NewSandbox(tarpit)

	// 5. Alert Dispatcher
	dispatcher := alert.NewDispatcher(&cfg.Alerts)

	// 6. Drift Monitor & Out-of-Band Callback Trap
	drift := canary.NewDriftMonitor(seeder, memStore)
	oob := canary.NewOOBManager(tokenMgr, func(hit *canary.OOBHit) {
		inc := tripwire.RecordOOBHit(hit.ClientIP, hit.UserAgent, hit)
		dispatcher.Dispatch(context.Background(), inc)
	})

	// 7. SSE Broker & API Handler
	sseBroker := NewSSEBroker(tripwire)
	apiHandler := NewAPIHandler(seeder, drift, oob, scanner, tripwire, interceptor, sandbox, dispatcher, sseBroker)

	// Set up Management HTTP Server
	apiMux := http.NewServeMux()
	apiHandler.RegisterRoutes(apiMux)
	apiMux.Handle("/", SPAHandler())

	apiServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      apiMux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Set up Proxy Gateway Server
	proxyGateway := NewProxyGateway(cfg.Server.UpstreamURL, scanner, tripwire, interceptor, sandbox, dispatcher)
	proxyServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.ProxyPort),
		Handler:      proxyGateway,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &Server{
		cfg:         cfg,
		apiServer:   apiServer,
		proxyServer: proxyServer,
		seeder:      seeder,
		tripwire:    tripwire,
	}, nil
}

// Start boots the management server and proxy gateway, executing auto-seed if enabled.
func (s *Server) Start(ctx context.Context) error {
	// Auto-seed configured vector stores
	if s.cfg.Canary.AutoSeed {
		log.Printf("[HONEYAGENT] Auto-seeding canary lures across configured vector stores: %v", s.cfg.Canary.TargetStores)
		results, err := s.seeder.SeedAll(ctx)
		if err != nil {
			log.Printf("[HONEYAGENT] Warning during auto-seed: %v", err)
		} else {
			log.Printf("[HONEYAGENT] Successfully seeded %d canary documents into vector stores", len(results))
		}
	}

	errCh := make(chan error, 2)

	// Start Management Studio API
	go func() {
		log.Printf("[HONEYAGENT] Deception Studio & API running on http://%s", s.apiServer.Addr)
		if err := s.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("management server error: %w", err)
		}
	}()

	// Start Airlock Proxy Gateway
	go func() {
		log.Printf("[HONEYAGENT] Airlock Proxy Gateway running on http://%s -> forwarding to %s", s.proxyServer.Addr, s.cfg.Server.UpstreamURL)
		if err := s.proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("proxy gateway error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Println("[HONEYAGENT] Shutting down servers gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = s.apiServer.Shutdown(shutdownCtx)
		_ = s.proxyServer.Shutdown(shutdownCtx)
		return nil
	}
}
