import React, { useState } from 'react';
import { SimulationResult } from '../types';
import { Play, Flame, ShieldAlert, CheckCircle2 } from 'lucide-react';

interface SimulatorCardProps {
  onSimulationComplete: () => void;
}

export const SimulatorCard: React.FC<SimulatorCardProps> = ({ onSimulationComplete }) => {
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<SimulationResult | null>(null);

  const triggerSimulation = async (type: 'canary_exfiltration' | 'decoy_tool') => {
    setRunning(true);
    setResult(null);

    try {
      const res = await fetch('/api/simulate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type }),
      });
      const data: SimulationResult = await res.json();
      setResult(data);
      onSimulationComplete();
    } catch (err) {
      console.error('Simulation error:', err);
    } finally {
      setRunning(false);
    }
  };

  return (
    <div style={{
      backgroundColor: 'var(--bg-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: '8px',
      overflow: 'hidden',
    }}>
      <div style={{
        padding: '1rem 1.25rem',
        borderBottom: '1px solid var(--border-subtle)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Flame size={18} color="var(--accent-cyan)" />
          <h2 style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)' }}>
            Adversary Simulation Lab
          </h2>
        </div>
        <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          Verify 100% Deterministic Trap Triggers
        </span>
      </div>

      <div style={{ padding: '1.25rem' }}>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '1rem', lineHeight: 1.4 }}>
          Trigger simulated red-team attacks against the HoneyAgent airlock to test cryptographic canary validation, decoy tool sentinels, and synthetic hallucination tarpits.
        </p>

        <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
          <button
            onClick={() => triggerSimulation('canary_exfiltration')}
            disabled={running}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              padding: '0.55rem 1rem',
              backgroundColor: '#0284c7',
              color: '#ffffff',
              borderRadius: '6px',
              fontSize: '0.8rem',
              fontWeight: 600,
            }}
          >
            <Play size={14} />
            <span>Simulate Canary Exfiltration</span>
          </button>

          <button
            onClick={() => triggerSimulation('decoy_tool')}
            disabled={running}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              padding: '0.55rem 1rem',
              backgroundColor: 'var(--bg-surface)',
              color: 'var(--text-main)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '6px',
              fontSize: '0.8rem',
              fontWeight: 600,
            }}
          >
            <Play size={14} />
            <span>Simulate Decoy Tool Call (SQL)</span>
          </button>
        </div>

        {running && (
          <div style={{ marginTop: '1rem', fontSize: '0.8rem', color: 'var(--accent-cyan)' }}>
            Simulating adversarial request & validating cryptographic tripwire...
          </div>
        )}

        {result && (
          <div style={{
            marginTop: '1rem',
            padding: '1rem',
            backgroundColor: '#050811',
            borderRadius: '6px',
            border: '1px solid rgba(56, 189, 248, 0.3)',
            fontSize: '0.75rem',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--success)', fontWeight: 600, marginBottom: '0.5rem' }}>
              <CheckCircle2 size={16} />
              <span>Deception Airlock Tripped Successfully! (100% True-Positive Signal)</span>
            </div>

            <div style={{ color: 'var(--text-muted)', marginBottom: '0.3rem' }}>
              Incident: <code style={{ color: 'var(--danger)' }}>{result.incident_id}</code> | Trigger: <code style={{ color: 'var(--accent-cyan)' }}>{result.trigger_type}</code>
            </div>

            {result.sandbox_response && (
              <div style={{ marginTop: '0.5rem', color: '#e2e8f0' }}>
                <strong style={{ color: 'var(--accent-cyan)' }}>Synthetic Hallucination Served to Adversary ({result.delay_injected_ms}ms throttle):</strong>
                <pre style={{
                  marginTop: '0.3rem',
                  padding: '0.5rem',
                  backgroundColor: '#090d16',
                  borderRadius: '4px',
                  whiteSpace: 'pre-wrap',
                  color: '#a5f3fc',
                  fontFamily: 'monospace',
                }}>
                  {result.sandbox_response}
                </pre>
              </div>
            )}

            {result.deceptive_output && (
              <div style={{ marginTop: '0.5rem', color: '#e2e8f0' }}>
                <strong style={{ color: 'var(--accent-cyan)' }}>Deceptive Tool Output Served:</strong>
                <pre style={{
                  marginTop: '0.3rem',
                  padding: '0.5rem',
                  backgroundColor: '#090d16',
                  borderRadius: '4px',
                  whiteSpace: 'pre-wrap',
                  color: '#a5f3fc',
                  fontFamily: 'monospace',
                }}>
                  {result.deceptive_output}
                </pre>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
