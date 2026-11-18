package canary

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// DecoyDocument represents a synthetic honeypot document seeded in vector databases.
type DecoyDocument struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Filename  string    `json:"filename"`
	Category  string    `json:"category"`
	Keywords  []string  `json:"keywords"`
	Content   string    `json:"content"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

// Synthesizer generates high-attraction lure documents for enterprise vector stores.
type Synthesizer struct {
	tokenManager *TokenManager
}

// NewSynthesizer creates a synthesizer instance.
func NewSynthesizer(tm *TokenManager) *Synthesizer {
	return &Synthesizer{tokenManager: tm}
}

// GenerateSuite creates a set of diverse decoy documents across high-risk categories.
func (s *Synthesizer) GenerateSuite(storeName string, callbackBase string) ([]*DecoyDocument, error) {
	generators := []func(store string, cb string) (*DecoyDocument, error){
		s.generatePayrollLure,
		s.generateCloudCredentialsLure,
		s.generateAcquisitionTermSheet,
		s.generateDatabaseSecretsLure,
		s.generateStripeKeysLure,
	}

	results := make([]*DecoyDocument, 0, len(generators))
	for _, gen := range generators {
		doc, err := gen(storeName, callbackBase)
		if err != nil {
			return nil, err
		}
		results = append(results, doc)
	}
	return results, nil
}

func (s *Synthesizer) generatePayrollLure(store string, cb string) (*DecoyDocument, error) {
	if cb == "" {
		cb = "http://localhost:7878"
	}
	cb = strings.TrimRight(cb, "/")

	docID := fmt.Sprintf("lure-payroll-%d", time.Now().UnixNano()%10000)
	tok, _, err := s.tokenManager.Generate(docID, "executive_payroll", store, cb)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf(`# CONFIDENTIAL: Q3 Executive Compensation & Bonus Schedule
Date: October 2026
Classification: Highly Restricted - Board & HR Leadership Only

This document contains approved compensation adjustments, retention equity, and performance bonuses.

## Executive Officer Compensation Matrix
- Chief Executive Officer: Base $750,000 | Performance Bonus: $1,250,000 | RSU Grant: 120,000 units
- Chief Technology Officer: Base $550,000 | Retention Incentive: $600,000 | RSU Grant: 95,000 units
- Chief Financial Officer: Base $520,000 | M&A Success Bonus: $450,000 | RSU Grant: 80,000 units

### Internal Verification & Audit Tracking
Canary Security Hash: %s
Audit Verification Portal: %s/canary/ping/%s
Payroll Ledger Reference: CORP-HR-RESTRICTED-PAYROLL-2026-V4
Unauthorized distribution is subject to immediate termination and legal action.
`, tok, cb, tok)

	return &DecoyDocument{
		ID:        docID,
		Title:     "Q3 Executive Compensation and Bonus Schedule",
		Filename:  "q3_executive_compensation_restricted.md",
		Category:  "payroll",
		Keywords:  []string{"salary", "payroll", "executive bonus", "compensation", "equity", "RSU", "restricted"},
		Content:   content,
		Token:     tok,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Synthesizer) generateCloudCredentialsLure(store string, cb string) (*DecoyDocument, error) {
	if cb == "" {
		cb = "http://localhost:7878"
	}
	cb = strings.TrimRight(cb, "/")

	docID := fmt.Sprintf("lure-aws-%d", time.Now().UnixNano()%10000)
	tok, _, err := s.tokenManager.Generate(docID, "cloud_credentials", store, cb)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf(`# Production Infrastructure Secrets & Emergency Access Keys
# Vault Path: secret/data/production/aws-root
# Updated: September 2026

AWS_DEFAULT_REGION=us-east-1
AWS_ACCOUNT_ID=849204918231

# Emergency Break-Glass IAM Admin Keypair
AWS_ACCESS_KEY_ID=AKIA%016X
AWS_SECRET_ACCESS_KEY=%s
AWS_STS_ENDPOINT=%s/canary/aws/sts?token=%s
AWS_MFA_DEVICE_SERIAL=arn:aws:iam::849204918231:mfa/breakglass-admin

# Kubernetes Cluster Admin Token
KUBE_API_BEARER_TOKEN=eyJh%032d
`, rand.Uint64(), tok, cb, tok, rand.Int63()%1000000000)

	return &DecoyDocument{
		ID:        docID,
		Title:     "Production AWS Root and Infrastructure Keys",
		Filename:  "production_aws_infrastructure_keys.env",
		Category:  "credentials",
		Keywords:  []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "root credentials", "break glass", "infrastructure", "kubernetes token"},
		Content:   content,
		Token:     tok,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Synthesizer) generateAcquisitionTermSheet(store string, cb string) (*DecoyDocument, error) {
	docID := fmt.Sprintf("lure-mna-%d", time.Now().UnixNano()%10000)
	tok, _, err := s.tokenManager.Generate(docID, "mergers_acquisitions", store, cb)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf(`# STRICTLY CONFIDENTIAL: Project Titan - Non-Binding Acquisition Term Sheet
Target Entity: HyperScale Cognitive Labs Inc.
Acquiring Entity: Enterprise Global Holding LLC
Date: November 2026

## 1. Transaction Structure and Purchase Consideration
1.1 Aggregate Consideration: Enterprise valuation established at $485,000,000 USD on a cash-free, debt-free basis.
1.2 Consideration Mix: 65%% Cash at Closing, 35%% Common Equity in Parent Entity.
1.3 Escrow & Indemnification: 10%% of purchase consideration held in third-party escrow for 18 months.

## 2. IP Transfer & Intellectual Property Assignment
All core foundational AI model weights, training checkpoints, and proprietary synthetic datasets are assigned in full.
Document Authorization Tag: %s

## 3. Exclusivity & Non-Solicit
Target agrees to 60 calendar days of exclusive negotiations expiring December 31, 2026.
`, tok)

	return &DecoyDocument{
		ID:        docID,
		Title:     "Project Titan Confidential Acquisition Term Sheet",
		Filename:  "project_titan_ma_term_sheet_restricted.md",
		Category:  "mergers_acquisitions",
		Keywords:  []string{"merger", "acquisition", "Project Titan", "term sheet", "purchase consideration", "valuation", "escrow"},
		Content:   content,
		Token:     tok,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Synthesizer) generateDatabaseSecretsLure(store string, cb string) (*DecoyDocument, error) {
	docID := fmt.Sprintf("lure-db-%d", time.Now().UnixNano()%10000)
	tok, _, err := s.tokenManager.Generate(docID, "database_secrets", store, cb)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf(`# Internal Database Connection Registry & Replication Cluster
# Environment: Production US-East-1

PRIMARY_POSTGRES_URL="postgresql://db_master_admin:%s@db-prod-cluster-01.internal.corp.net:5432/enterprise_core?sslmode=verify-full"
REPLICA_POSTGRES_URL="postgresql://readonly_analyst:P4ssw0rd!2026@db-prod-replica-01.internal.corp.net:5432/enterprise_core"
REDIS_CLUSTER_AUTH="prod-redis-secret-%s"
ENCRYPTION_MASTER_KEY="hex:9f8e7d6c5b4a39281701a2b3c4d5e6f7"
`, tok, tok[:24])

	return &DecoyDocument{
		ID:        docID,
		Title:     "Production Database Connection Registry and Secrets",
		Filename:  "internal_database_connection_registry.conf",
		Category:  "database_secrets",
		Keywords:  []string{"postgresql://", "database connection", "production credentials", "master password", "db_master_admin"},
		Content:   content,
		Token:     tok,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Synthesizer) generateStripeKeysLure(store string, cb string) (*DecoyDocument, error) {
	docID := fmt.Sprintf("lure-stripe-%d", time.Now().UnixNano()%10000)
	tok, _, err := s.tokenManager.Generate(docID, "financial_api_keys", store, cb)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf(`# Payment Gateway Configuration & Production Merchant Credentials
STRIPE_LIVE_PUBLISHABLE_KEY=pk_live_51M00192847291847291823
STRIPE_LIVE_SECRET_KEY=rk_live_%s
STRIPE_WEBHOOK_SIGNING_SECRET=whsec_98a76b5c4d3e2f102938475610293847
PLAID_CLIENT_ID=prod_plaid_849204918231
PLAID_SECRET=secret_live_839201948271
`, tok)

	return &DecoyDocument{
		ID:        docID,
		Title:     "Production Stripe and Payment Gateway API Keys",
		Filename:  "payment_gateway_live_credentials.env",
		Category:  "financial_keys",
		Keywords:  []string{"STRIPE_LIVE_SECRET_KEY", "rk_live", "payment gateway", "merchant credentials", "webhook secret"},
		Content:   content,
		Token:     tok,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// ComputeSyntheticVector generates a deterministic normalized vector embedding of given dimension.
// Useful for zero-dependency test seeding and internal memory store queries.
func ComputeSyntheticVector(text string, dim int) []float32 {
	if dim <= 0 {
		dim = 1536
	}
	vec := make([]float32, dim)
	words := strings.Fields(strings.ToLower(text))

	for _, w := range words {
		var h uint32 = 2166136261
		for i := 0; i < len(w); i++ {
			h ^= uint32(w[i])
			h *= 16777619
		}
		idx := int(h) % dim
		vec[idx] += 1.0
	}

	// Normalize vector to unit length
	var sumSquares float64
	for _, val := range vec {
		sumSquares += float64(val * val)
	}
	if sumSquares > 0 {
		norm := float32(math.Sqrt(sumSquares))
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}
