package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/config"
	"github.com/alexandrmotologa/honeyagent/internal/detector"
)

// Dispatcher fan-outs security incidents to configured SIEM and notification sinks.
type Dispatcher struct {
	cfg        *config.AlertConfig
	siem       *SIEMClient
	slack      *SlackClient
	pagerDuty  *PagerDutyClient
	httpClient *http.Client
}

// NewDispatcher initializes the multi-channel alert dispatcher.
func NewDispatcher(cfg *config.AlertConfig) *Dispatcher {
	return &Dispatcher{
		cfg:       cfg,
		siem:      NewSIEMClient(&cfg.SyslogCEF),
		slack:     NewSlackClient(&cfg.Slack),
		pagerDuty: NewPagerDutyClient(&cfg.PagerDuty),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Dispatch sends the incident concurrently to all enabled channels.
func (d *Dispatcher) Dispatch(ctx context.Context, inc *detector.Incident) {
	var wg sync.WaitGroup

	// 1. SIEM CEF/Syslog
	if d.cfg.SyslogCEF.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := d.siem.Send(inc); err != nil {
				log.Printf("[ALERT] SIEM dispatch error: %v", err)
			}
		}()
	}

	// 2. Slack Webhook
	if d.cfg.Slack.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := d.slack.Send(c, inc); err != nil {
				log.Printf("[ALERT] Slack dispatch error: %v", err)
			}
		}()
	}

	// 3. PagerDuty
	if d.cfg.PagerDuty.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := d.pagerDuty.Send(c, inc); err != nil {
				log.Printf("[ALERT] PagerDuty dispatch error: %v", err)
			}
		}()
	}

	// 4. Custom Generic Webhook
	if d.cfg.WebhookURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			d.sendWebhook(c, inc)
		}()
	}

	// Wait or background depending on caller context
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		// Non-blocking timeout guard
	}
}

func (d *Dispatcher) sendWebhook(ctx context.Context, inc *detector.Incident) {
	data, err := json.Marshal(inc)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.cfg.WebhookURL, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HoneyAgent-Airlock/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		log.Printf("[ALERT] Custom webhook error: %v", err)
		return
	}
	defer resp.Body.Close()
}
