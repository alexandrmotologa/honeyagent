import React from 'react';
import { Incident } from '../types';
import { ShieldAlert, ExternalLink, Terminal, Key } from 'lucide-react';

interface ThreatFeedProps {
  incidents: Incident[];
  onSelectIncident: (inc: Incident) => void;
}

export const ThreatFeed: React.FC<ThreatFeedProps> = ({ incidents, onSelectIncident }) => {
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
        justifyContent: 'space-between',
        alignItems: 'center',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <ShieldAlert size={18} color="var(--danger)" />
          <h2 style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)' }}>
            Live Adversary Incidents Feed
          </h2>
        </div>
        <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          {incidents.length} verified detections
        </span>
      </div>

      <div style={{ maxHeight: '450px', overflowY: 'auto' }}>
        {incidents.length === 0 ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
            No security incidents recorded. System airlock active and monitoring.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            {incidents.map((inc) => (
              <div
                key={inc.id}
                style={{
                  padding: '1rem 1.25rem',
                  borderBottom: '1px solid var(--border-subtle)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: '1rem',
                  transition: 'background-color 0.15s',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.75rem', flex: 1 }}>
                  <div style={{
                    padding: '0.35rem',
                    borderRadius: '6px',
                    backgroundColor: inc.trigger_type === 'canary_token' ? 'rgba(56, 189, 248, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                    color: inc.trigger_type === 'canary_token' ? 'var(--accent-cyan)' : 'var(--danger)',
                    marginTop: '0.15rem',
                  }}>
                    {inc.trigger_type === 'canary_token' ? <Key size={16} /> : <Terminal size={16} />}
                  </div>

                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                      <span style={{
                        fontSize: '0.65rem',
                        fontWeight: 700,
                        padding: '0.15rem 0.4rem',
                        borderRadius: '4px',
                        backgroundColor: 'var(--danger-bg)',
                        color: 'var(--danger)',
                        border: '1px solid rgba(239, 68, 68, 0.3)',
                      }}>
                        {inc.severity.toUpperCase()}
                      </span>
                      <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)' }}>
                        {inc.title}
                      </span>
                    </div>

                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
                      {inc.summary}
                    </div>

                    <div style={{ display: 'flex', gap: '1rem', marginTop: '0.35rem', fontSize: '0.7rem', color: 'var(--text-muted)' }}>
                      <span>IP: <code style={{ color: 'var(--accent-cyan)' }}>{inc.client_ip}</code></span>
                      <span>Session: <code>{inc.session_id}</code></span>
                      <span>{new Date(inc.timestamp).toLocaleTimeString()}</span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={() => onSelectIncident(inc)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.35rem',
                    padding: '0.4rem 0.75rem',
                    backgroundColor: 'var(--bg-surface)',
                    color: 'var(--accent-cyan)',
                    border: '1px solid var(--border-subtle)',
                    borderRadius: '6px',
                    fontSize: '0.75rem',
                    fontWeight: 500,
                    whiteSpace: 'nowrap',
                  }}
                >
                  <span>Replay Trajectory</span>
                  <ExternalLink size={12} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
