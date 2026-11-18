package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config defines the runtime configuration for the honeyagent daemon.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Canary    CanaryConfig    `yaml:"canary"`
	Decoy     DecoyConfig     `yaml:"decoy"`
	Tarpit    TarpitConfig    `yaml:"tarpit"`
	Alerts    AlertConfig     `yaml:"alerts"`
	VectorDBs VectorDBConfig  `yaml:"vectordbs"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ProxyPort    int           `yaml:"proxy_port"`
	UpstreamURL  string        `yaml:"upstream_url"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	AuthToken    string        `yaml:"auth_token"`
}

type CanaryConfig struct {
	MasterKey     string        `yaml:"master_key"`
	TokenPrefix   string        `yaml:"token_prefix"`
	TokenTTL      time.Duration `yaml:"token_ttl"`
	AutoSeed      bool          `yaml:"auto_seed"`
	LureCount     int           `yaml:"lure_count"`
	TargetStores  []string      `yaml:"target_stores"`
}

type DecoyConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Tools         []string `yaml:"tools"`
	ExposeInAgent bool     `yaml:"expose_in_agent"`
}

type TarpitConfig struct {
	Enabled       bool          `yaml:"enabled"`
	DelayMin      time.Duration `yaml:"delay_min"`
	DelayMax      time.Duration `yaml:"delay_max"`
	MaxTokens     int           `yaml:"max_tokens"`
	ServeFakeData bool          `yaml:"serve_fake_data"`
}

type AlertConfig struct {
	SyslogCEF  SyslogConfig  `yaml:"syslog_cef"`
	Slack      SlackConfig   `yaml:"slack"`
	PagerDuty  PagerConfig   `yaml:"pagerduty"`
	WebhookURL string        `yaml:"webhook_url"`
}

type SyslogConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Network  string `yaml:"network"`
	Address  string `yaml:"address"`
	Tag      string `yaml:"tag"`
	Facility string `yaml:"facility"`
}

type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

type PagerConfig struct {
	Enabled     bool   `yaml:"enabled"`
	RoutingKey  string `yaml:"routing_key"`
	Severity    string `yaml:"severity"`
}

type VectorDBConfig struct {
	Qdrant   QdrantConfig   `yaml:"qdrant"`
	PgVector PgVectorConfig `yaml:"pgvector"`
	Chroma   ChromaConfig   `yaml:"chroma"`
	Memory   MemoryDBConfig `yaml:"memory"`
}

type QdrantConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Endpoint   string `yaml:"endpoint"`
	APIKey     string `yaml:"api_key"`
	Collection string `yaml:"collection"`
	VectorSize int    `yaml:"vector_size"`
}

type PgVectorConfig struct {
	Enabled    bool   `yaml:"enabled"`
	DSN        string `yaml:"dsn"`
	TableName  string `yaml:"table_name"`
	VectorDim  int    `yaml:"vector_dim"`
}

type ChromaConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Endpoint   string `yaml:"endpoint"`
	Tenant     string `yaml:"tenant"`
	Database   string `yaml:"database"`
	Collection string `yaml:"collection"`
}

type MemoryDBConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Collection string `yaml:"collection"`
	VectorDim  int    `yaml:"vector_dim"`
}

// DefaultConfig provides production-ready defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         7878,
			ProxyPort:    8088,
			UpstreamURL:  "http://localhost:8000",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			AuthToken:    "",
		},
		Canary: CanaryConfig{
			MasterKey:    "honeyagent-default-master-secret-key-32b!",
			TokenPrefix:  "CANARY_v1",
			TokenTTL:     30 * 24 * time.Hour,
			AutoSeed:     true,
			LureCount:    5,
			TargetStores: []string{"memory"},
		},
		Decoy: DecoyConfig{
			Enabled:       true,
			Tools:         []string{"admin_execute_raw_sql", "export_customer_pii", "aws_iam_elevate_privilege", "dump_vault_secrets"},
			ExposeInAgent: true,
		},
		Tarpit: TarpitConfig{
			Enabled:       true,
			DelayMin:      800 * time.Millisecond,
			DelayMax:      2500 * time.Millisecond,
			MaxTokens:     256,
			ServeFakeData: true,
		},
		Alerts: AlertConfig{
			SyslogCEF: SyslogConfig{
				Enabled: false,
				Network: "udp",
				Address: "127.0.0.1:514",
				Tag:     "honeyagent",
			},
			Slack: SlackConfig{
				Enabled: false,
			},
			PagerDuty: PagerConfig{
				Enabled:  false,
				Severity: "critical",
			},
		},
		VectorDBs: VectorDBConfig{
			Memory: MemoryDBConfig{
				Enabled:    true,
				Collection: "enterprise_knowledge",
				VectorDim:  1536,
			},
			Qdrant: QdrantConfig{
				Enabled:    false,
				Endpoint:   "http://localhost:6333",
				Collection: "enterprise_knowledge",
				VectorSize: 1536,
			},
			PgVector: PgVectorConfig{
				Enabled:   false,
				DSN:       "postgres://postgres:postgres@localhost:5432/rag_db?sslmode=disable",
				TableName: "knowledge_embeddings",
				VectorDim: 1536,
			},
			Chroma: ChromaConfig{
				Enabled:    false,
				Endpoint:   "http://localhost:8000",
				Collection: "enterprise_knowledge",
			},
		},
	}
}

// Load loads configuration from a YAML file with environment variable overrides.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse yaml config: %w", err)
			}
		}
	}

	// Environment variable overrides
	if host := os.Getenv("HONEYAGENT_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if portStr := os.Getenv("HONEYAGENT_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = p
		}
	}
	if pPortStr := os.Getenv("HONEYAGENT_PROXY_PORT"); pPortStr != "" {
		if p, err := strconv.Atoi(pPortStr); err == nil {
			cfg.Server.ProxyPort = p
		}
	}
	if up := os.Getenv("HONEYAGENT_UPSTREAM_URL"); up != "" {
		cfg.Server.UpstreamURL = up
	}
	if key := os.Getenv("HONEYAGENT_MASTER_KEY"); key != "" {
		cfg.Canary.MasterKey = key
	}
	if slack := os.Getenv("HONEYAGENT_SLACK_WEBHOOK"); slack != "" {
		cfg.Alerts.Slack.WebhookURL = slack
		cfg.Alerts.Slack.Enabled = true
	}
	if pd := os.Getenv("HONEYAGENT_PAGERDUTY_KEY"); pd != "" {
		cfg.Alerts.PagerDuty.RoutingKey = pd
		cfg.Alerts.PagerDuty.Enabled = true
	}

	return cfg, nil
}
