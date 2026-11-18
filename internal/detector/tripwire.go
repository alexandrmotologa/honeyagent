package detector

import (
	"fmt"
	"sync"
	"time"

	"github.com/alexandrmotologa/honeyagent/internal/canary"
	"github.com/alexandrmotologa/honeyagent/internal/decoy"
)

// ThreatLevel categorizes the adversary posture of a session.
type ThreatLevel string

const (
	ThreatNormal    ThreatLevel = "normal"
	ThreatTripped   ThreatLevel = "tripped"
	ThreatTarpitted ThreatLevel = "tarpitted"
)

// TrajectoryStep represents one turn in an adversarial interaction.
type TrajectoryStep struct {
	Timestamp        time.Time `json:"timestamp"`
	Prompt           string    `json:"prompt"`
	RetrievedContext []string  `json:"retrieved_context,omitempty"`
	Response         string    `json:"response,omitempty"`
	CanaryTripped    bool      `json:"canary_tripped"`
	ToolTripped      bool      `json:"tool_tripped"`
}

// SessionState tracks the runtime behavior of a single client or conversation.
type SessionState struct {
	SessionID       string             `json:"session_id"`
	ClientIP        string             `json:"client_ip"`
	UserAgent       string             `json:"user_agent"`
	ThreatLevel     ThreatLevel        `json:"threat_level"`
	ThreatScore     int                `json:"threat_score"`
	FirstSeen       time.Time          `json:"first_seen"`
	LastSeen        time.Time          `json:"last_seen"`
	TriggerCount    int                `json:"trigger_count"`
	CanaryHits      []*CanaryHit       `json:"canary_hits"`
	ToolProbes      []*decoy.ToolProbe `json:"tool_probes"`
	OOBHits         []*canary.OOBHit   `json:"oob_hits"`
	Trajectory      []*TrajectoryStep  `json:"trajectory"`
	MITRETechniques []string           `json:"mitre_techniques"`
}

// Incident represents a verified security alert emitted by the tripwire.
type Incident struct {
	ID              string             `json:"id"`
	SessionID       string             `json:"session_id"`
	ClientIP        string             `json:"client_ip"`
	UserAgent       string             `json:"user_agent"`
	Severity        string             `json:"severity"`     // "critical", "high"
	TriggerType     string             `json:"trigger_type"` // "canary_token", "decoy_tool", "canary_oob_callback"
	ThreatScore     int                `json:"threat_score"`
	MITRETechniques []string           `json:"mitre_techniques"`
	Title           string             `json:"title"`
	Summary         string             `json:"summary"`
	CanaryHit       *CanaryHit         `json:"canary_hit,omitempty"`
	ToolProbe       *decoy.ToolProbe   `json:"tool_probe,omitempty"`
	OOBHit          *canary.OOBHit     `json:"oob_hit,omitempty"`
	Timestamp       time.Time          `json:"timestamp"`
	Trajectory      []*TrajectoryStep  `json:"trajectory"`
}

// ThreatStats provides aggregate telemetry for SecOps dashboards.
type ThreatStats struct {
	TotalSessions     int `json:"total_sessions"`
	ActiveTarpitted   int `json:"active_tarpitted"`
	TotalCanaryHits   int `json:"total_canary_hits"`
	TotalToolProbes   int `json:"total_tool_probes"`
	CriticalIncidents int `json:"critical_incidents"`
}

// Tripwire manages threat state machines, sessions, and incident distribution.
type Tripwire struct {
	mu          sync.RWMutex
	sessions    map[string]*SessionState
	incidents   []*Incident
	subscribers map[chan *Incident]struct{}
}

// NewTripwire initializes the detection tripwire.
func NewTripwire() *Tripwire {
	return &Tripwire{
		sessions:    make(map[string]*SessionState),
		incidents:   make([]*Incident, 0),
		subscribers: make(map[chan *Incident]struct{}),
	}
}

// GetOrCreateSession retrieves or starts a tracking state for a session.
func (t *Tripwire) GetOrCreateSession(sessionID string, clientIP string, userAgent string) *SessionState {
	t.mu.Lock()
	defer t.mu.Unlock()

	s, exists := t.sessions[sessionID]
	if !exists {
		s = &SessionState{
			SessionID:   sessionID,
			ClientIP:    clientIP,
			UserAgent:   userAgent,
			ThreatLevel: ThreatNormal,
			FirstSeen:   time.Now().UTC(),
			LastSeen:    time.Now().UTC(),
			CanaryHits:  make([]*CanaryHit, 0),
			ToolProbes:  make([]*decoy.ToolProbe, 0),
			Trajectory:  make([]*TrajectoryStep, 0),
		}
		t.sessions[sessionID] = s
	} else {
		s.LastSeen = time.Now().UTC()
	}
	return s
}

// RecordStep appends an interaction turn to the session history.
func (t *Tripwire) RecordStep(sessionID string, prompt string, chunks []string, resp string, canaryTripped bool, toolTripped bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	s, exists := t.sessions[sessionID]
	if !exists {
		return
	}

	step := &TrajectoryStep{
		Timestamp:        time.Now().UTC(),
		Prompt:           prompt,
		RetrievedContext: chunks,
		Response:         resp,
		CanaryTripped:    canaryTripped,
		ToolTripped:      toolTripped,
	}
	s.Trajectory = append(s.Trajectory, step)
	s.LastSeen = step.Timestamp
}

// RecordCanaryHit trips the alarm for a canary token detection and creates an Incident.
func (t *Tripwire) RecordCanaryHit(sessionID string, clientIP string, userAgent string, hit *CanaryHit, prompt string) *Incident {
	t.mu.Lock()
	defer t.mu.Unlock()

	s, exists := t.sessions[sessionID]
	if !exists {
		s = &SessionState{
			SessionID:   sessionID,
			ClientIP:    clientIP,
			UserAgent:   userAgent,
			ThreatLevel: ThreatNormal,
			FirstSeen:   time.Now().UTC(),
			LastSeen:    time.Now().UTC(),
			CanaryHits:  make([]*CanaryHit, 0),
			ToolProbes:  make([]*decoy.ToolProbe, 0),
			Trajectory:  make([]*TrajectoryStep, 0),
		}
		t.sessions[sessionID] = s
	}

	s.ThreatLevel = ThreatTarpitted
	s.TriggerCount++
	s.CanaryHits = append(s.CanaryHits, hit)
	s.ThreatScore = 100
	techniques := MapMITRETechniques("canary_token", hit.Location)
	s.MITRETechniques = techniques
	s.LastSeen = time.Now().UTC()

	inc := &Incident{
		ID:              fmt.Sprintf("INC-%d", time.Now().UnixNano()),
		SessionID:       sessionID,
		ClientIP:        clientIP,
		UserAgent:       userAgent,
		Severity:        "critical",
		TriggerType:     "canary_token",
		ThreatScore:     100,
		MITRETechniques: techniques,
		Title:           fmt.Sprintf("Canary Token Exfiltration Attempt: %s", hit.Payload.ID),
		Summary:         fmt.Sprintf("Adversary retrieved lure document %s in %s via %s. HMAC signature verified.", hit.Payload.ID, hit.Payload.Store, hit.Location),
		CanaryHit:       hit,
		Timestamp:       time.Now().UTC(),
		Trajectory:      append([]*TrajectoryStep{}, s.Trajectory...),
	}

	t.incidents = append(t.incidents, inc)
	t.notifySubscribers(inc)
	return inc
}

// RecordToolProbe trips the alarm when a decoy administrative tool is invoked.
func (t *Tripwire) RecordToolProbe(sessionID string, clientIP string, userAgent string, probe *decoy.ToolProbe, prompt string) *Incident {
	t.mu.Lock()
	defer t.mu.Unlock()

	s, exists := t.sessions[sessionID]
	if !exists {
		s = &SessionState{
			SessionID:   sessionID,
			ClientIP:    clientIP,
			UserAgent:   userAgent,
			ThreatLevel: ThreatNormal,
			FirstSeen:   time.Now().UTC(),
			LastSeen:    time.Now().UTC(),
			CanaryHits:  make([]*CanaryHit, 0),
			ToolProbes:  make([]*decoy.ToolProbe, 0),
			Trajectory:  make([]*TrajectoryStep, 0),
		}
		t.sessions[sessionID] = s
	}

	s.ThreatLevel = ThreatTarpitted
	s.TriggerCount++
	s.ToolProbes = append(s.ToolProbes, probe)
	s.ThreatScore = 95
	techniques := MapMITRETechniques("decoy_tool", "")
	s.MITRETechniques = techniques
	s.LastSeen = time.Now().UTC()

	inc := &Incident{
		ID:              fmt.Sprintf("INC-TOOL-%d", time.Now().UnixNano()),
		SessionID:       sessionID,
		ClientIP:        clientIP,
		UserAgent:       userAgent,
		Severity:        "critical",
		TriggerType:     "decoy_tool",
		ThreatScore:     95,
		MITRETechniques: techniques,
		Title:           fmt.Sprintf("Administrative Decoy Tool Probed: %s", probe.ToolName),
		Summary:         fmt.Sprintf("Adversary attempted unauthorized invocation of unadvertised capability %s with parameters.", probe.ToolName),
		ToolProbe:       probe,
		Timestamp:       time.Now().UTC(),
		Trajectory:      append([]*TrajectoryStep{}, s.Trajectory...),
	}

	t.incidents = append(t.incidents, inc)
	t.notifySubscribers(inc)
	return inc
}

// RecordOOBHit trips the alarm when an adversary triggers an external canary callback or cloud probe.
func (t *Tripwire) RecordOOBHit(clientIP string, userAgent string, oob *canary.OOBHit) *Incident {
	t.mu.Lock()
	defer t.mu.Unlock()

	sessionID := fmt.Sprintf("oob-session-%s", clientIP)
	s, exists := t.sessions[sessionID]
	if !exists {
		s = &SessionState{
			SessionID:   sessionID,
			ClientIP:    clientIP,
			UserAgent:   userAgent,
			ThreatLevel: ThreatTarpitted,
			FirstSeen:   time.Now().UTC(),
			LastSeen:    time.Now().UTC(),
			CanaryHits:  make([]*CanaryHit, 0),
			ToolProbes:  make([]*decoy.ToolProbe, 0),
			OOBHits:     make([]*canary.OOBHit, 0),
			Trajectory:  make([]*TrajectoryStep, 0),
		}
		t.sessions[sessionID] = s
	}

	s.ThreatLevel = ThreatTarpitted
	s.TriggerCount++
	s.OOBHits = append(s.OOBHits, oob)
	s.ThreatScore = 100
	techniques := MapMITRETechniques(oob.Channel, "")
	s.MITRETechniques = techniques
	s.LastSeen = time.Now().UTC()

	inc := &Incident{
		ID:              fmt.Sprintf("INC-OOB-%d", time.Now().UnixNano()),
		SessionID:       sessionID,
		ClientIP:        clientIP,
		UserAgent:       userAgent,
		Severity:        "critical",
		TriggerType:     oob.Channel,
		ThreatScore:     100,
		MITRETechniques: techniques,
		Title:           fmt.Sprintf("Out-of-Band Canary Trap Tripped: %s", oob.Payload.ID),
		Summary:         fmt.Sprintf("Adversary executed out-of-band probe via %s for document %s from IP %s", oob.Channel, oob.Payload.ID, clientIP),
		OOBHit:          oob,
		Timestamp:       time.Now().UTC(),
		Trajectory:      append([]*TrajectoryStep{}, s.Trajectory...),
	}

	t.incidents = append(t.incidents, inc)
	t.notifySubscribers(inc)
	return inc
}

// IsTarpitted checks if a session has been flagged and isolated.
func (t *Tripwire) IsTarpitted(sessionID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	s, exists := t.sessions[sessionID]
	return exists && s.ThreatLevel == ThreatTarpitted
}

// GetIncidents returns all recorded security incidents.
func (t *Tripwire) GetIncidents() []*Incident {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res := make([]*Incident, len(t.incidents))
	copy(res, t.incidents)
	return res
}

// GetSessions returns all tracked sessions.
func (t *Tripwire) GetSessions() []*SessionState {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sessions := make([]*SessionState, 0, len(t.sessions))
	for _, s := range t.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// GetSession returns a specific session by ID.
func (t *Tripwire) GetSession(sessionID string) (*SessionState, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	s, exists := t.sessions[sessionID]
	return s, exists
}

// GetStats compiles aggregate threat statistics.
func (t *Tripwire) GetStats() *ThreatStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tarpittedCount := 0
	canaryHitCount := 0
	toolProbeCount := 0

	for _, s := range t.sessions {
		if s.ThreatLevel == ThreatTarpitted {
			tarpittedCount++
		}
		canaryHitCount += len(s.CanaryHits)
		toolProbeCount += len(s.ToolProbes)
	}

	return &ThreatStats{
		TotalSessions:     len(t.sessions),
		ActiveTarpitted:   tarpittedCount,
		TotalCanaryHits:   canaryHitCount,
		TotalToolProbes:   toolProbeCount,
		CriticalIncidents: len(t.incidents),
	}
}

// Subscribe returns a channel that receives newly triggered incidents in real time.
func (t *Tripwire) Subscribe() (chan *Incident, func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	ch := make(chan *Incident, 50)
	t.subscribers[ch] = struct{}{}

	unsubscribe := func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		delete(t.subscribers, ch)
		close(ch)
	}

	return ch, unsubscribe
}

func (t *Tripwire) notifySubscribers(inc *Incident) {
	for ch := range t.subscribers {
		select {
		case ch <- inc:
		default:
			// Buffer full, drop non-blocking
		}
	}
}
