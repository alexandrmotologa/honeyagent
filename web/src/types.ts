export interface ThreatStats {
  total_sessions: number;
  active_tarpitted: number;
  total_canary_hits: number;
  total_tool_probes: number;
  critical_incidents: number;
  seeded_canaries_count: number;
  quarantined_turns: number;
  delay_injected_seconds: number;
}

export interface TrajectoryStep {
  timestamp: string;
  prompt: string;
  retrieved_context?: string[];
  response?: string;
  canary_tripped: boolean;
  tool_tripped: boolean;
}

export interface CanaryHit {
  token: string;
  payload: {
    id: string;
    cat: string;
    st: string;
    ts: number;
    exp?: number;
    rnd: string;
  };
  location: string;
}

export interface ToolProbe {
  id: string;
  session_id: string;
  tool_name: string;
  arguments: Record<string, unknown>;
  client_ip: string;
  timestamp: string;
  risk_level: string;
}

export interface Incident {
  id: string;
  session_id: string;
  client_ip: string;
  user_agent: string;
  severity: string;
  trigger_type: string;
  title: string;
  summary: string;
  canary_hit?: CanaryHit;
  tool_probe?: ToolProbe;
  timestamp: string;
  trajectory?: TrajectoryStep[];
}

export interface SeedResult {
  document_id: string;
  store: string;
  status: string;
  token: string;
  title: string;
  category: string;
  error?: string;
  seeded_at: string;
}

export interface ToolStats {
  tool_name: string;
  probes_count: number;
  last_probed: string;
}

export interface SimulationResult {
  status: string;
  incident_id?: string;
  session_id: string;
  trigger_type: string;
  canary_token?: string;
  lure_document_id?: string;
  decoy_tool?: string;
  sandbox_response?: string;
  deceptive_output?: string;
  delay_injected_ms?: number;
}
