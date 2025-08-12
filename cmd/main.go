package main

import (
	"context"

	"log"
	"net/http"
	"time"

	"github.com/h3adex/stackit-exporter/internal/collector"
	"github.com/h3adex/stackit-exporter/internal/config"
	"github.com/h3adex/stackit-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sdkConfig "github.com/stackitcloud/stackit-sdk-go/core/config"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
	"github.com/stackitcloud/stackit-sdk-go/services/ske"
)

type Manager struct {
	ctx          context.Context
	iaasClient   *iaas.APIClient
	skeClient    *ske.APIClient
	projectID    string
	region       string
	iaasRegistry *metrics.IaasRegistry
	skeRegistry  *metrics.SKERegistry
}

func NewManager(ctx context.Context, projectID, region string, iaasClient *iaas.APIClient, skeClient *ske.APIClient) *Manager {
	return &Manager{
		ctx:          ctx,
		iaasClient:   iaasClient,
		skeClient:    skeClient,
		projectID:    projectID,
		region:       region,
		iaasRegistry: metrics.NewIaasRegistry(),
		skeRegistry:  metrics.NewSKERegistry(),
	}
}

func (m *Manager) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Scraping metrics...")
			collector.ScrapeIaasAPI(m.ctx, m.iaasClient, m.projectID, m.iaasRegistry)
			collector.ScrapeSkeAPI(m.ctx, m.skeClient, m.projectID, m.region, m.skeRegistry)
		case <-m.ctx.Done():
			log.Println("Stopping manager")
			return
		}
	}
}

func main() {
	ctx := context.Background()
	cfg, err := config.Load() // Ensure Load returns (*Config, error)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	iaasClientConfig := []sdkConfig.ConfigurationOption{
		sdkConfig.WithRegion(cfg.Region),
	}

	iaasClient, err := iaas.NewAPIClient(iaasClientConfig...)
	if err != nil {
		log.Fatalf("Error creating Iaas client: %v", err)
	}

	skeClient, err := ske.NewAPIClient()
	if err != nil {
		log.Fatalf("Error creating SKE client: %v", err)
	}

	manager := NewManager(ctx, cfg.ProjectID, cfg.Region, iaasClient, skeClient)

	go manager.Run(2 * time.Second)

	// Set up HTTP handlers for metrics and health checks
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Start the HTTP server with timeouts
	listenAddress := ":8080"
	server := &http.Server{
		Addr:         listenAddress,
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Listening on %s", listenAddress)
	log.Fatal(server.ListenAndServe())
}
