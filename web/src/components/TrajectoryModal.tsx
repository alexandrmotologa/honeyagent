import React from 'react';
import { Incident } from '../types';
import { X, ShieldAlert, Terminal, Clock, Lock, Download, FileText, Target } from 'lucide-react';

interface TrajectoryModalProps {
  incident: Incident | null;
  onClose: () => void;
}

export const TrajectoryModal: React.FC<TrajectoryModalProps> = ({ incident, onClose }) => {
  if (!incident) return null;

  const downloadExport = (format: 'markdown' | 'stix') => {
    window.open(`/api/threats/export?id=${incident.id}&format=${format}`, '_blank');
  };

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(5, 8, 17, 0.85)',
      backdropFilter: 'blur(4px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
      padding: '1.5rem',
    }}>
      <div style={{
        backgroundColor: 'var(--bg-card)',
        border: '1px solid var(--border-subtle)',
        borderRadius: '12px',
        width: '100%',
        maxWidth: '850px',
        maxHeight: '90vh',
        display: 'flex',
        flexDirection: 'column',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.7)',
        overflow: 'hidden',
      }}>
        {/* Modal Header */}
        <div style={{
          padding: '1.25rem 1.5rem',
          borderBottom: '1px solid var(--border-subtle)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div style={{ padding: '0.4rem', backgroundColor: 'var(--danger-bg)', borderRadius: '6px' }}>
              <ShieldAlert size={20} color="var(--danger)" />
            </div>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <h2 style={{ fontSize: '1.1rem', fontWeight: 600, color: 'var(--text-main)' }}>
                  Attacker Trajectory Replay: {incident.id}
                </h2>
                <span style={{
                  fontSize: '0.7rem',
                  fontWeight: 700,
                  padding: '0.15rem 0.4rem',
                  borderRadius: '4px',
                  backgroundColor: 'var(--danger-bg)',
                  color: 'var(--danger)',
                }}>
                  SCORE {incident.threat_score ?? 100}/100
                </span>
              </div>
              <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                Session: {incident.session_id} | Origin: {incident.client_ip}
              </span>
            </div>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              color: 'var(--text-muted)',
              padding: '0.25rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: '4px',
            }}
          >
            <X size={20} />
          </button>
        </div>

        {/* Modal Body */}
        <div style={{ padding: '1.5rem', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
          {/* Metadata Cards */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: '0.75rem',
            padding: '1rem',
            backgroundColor: 'var(--bg-surface)',
            borderRadius: '8px',
            fontSize: '0.8rem',
          }}>
            <div>
              <div style={{ color: 'var(--text-muted)' }}>Trigger Type</div>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginTop: '0.2rem' }}>
                {incident.trigger_type.toUpperCase()}
              </div>
            </div>
            <div>
              <div style={{ color: 'var(--text-muted)' }}>Client User-Agent</div>
              <div style={{ fontWeight: 600, color: 'var(--text-main)', marginTop: '0.2rem', textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }}>
                {incident.user_agent || 'Unknown'}
              </div>
            </div>
            <div>
              <div style={{ color: 'var(--text-muted)' }}>Trigger Timestamp</div>
              <div style={{ fontWeight: 600, color: 'var(--text-main)', marginTop: '0.2rem' }}>
                {new Date(incident.timestamp).toLocaleTimeString()}
              </div>
            </div>
          </div>

          {/* MITRE ATLAS Techniques */}
          {incident.mitre_techniques && incident.mitre_techniques.length > 0 && (
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.75rem 1rem',
              backgroundColor: 'rgba(56, 189, 248, 0.05)',
              border: '1px solid rgba(56, 189, 248, 0.2)',
              borderRadius: '6px',
              fontSize: '0.75rem',
              flexWrap: 'wrap',
            }}>
              <span style={{ color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: '0.3rem', fontWeight: 600 }}>
                <Target size={14} color="var(--accent-cyan)" />
                MITRE ATLAS:
              </span>
              {incident.mitre_techniques.map(tech => (
                <span key={tech} style={{
                  padding: '0.15rem 0.4rem',
                  borderRadius: '4px',
                  backgroundColor: 'rgba(56, 189, 248, 0.15)',
                  color: 'var(--accent-cyan)',
                  fontFamily: 'monospace',
                  fontWeight: 600,
                }}>
                  {tech}
                </span>
              ))}
            </div>
          )}

          {/* Trigger Details */}
          {incident.canary_hit && (
            <div style={{
              backgroundColor: 'rgba(2, 132, 199, 0.1)',
              border: '1px solid rgba(56, 189, 248, 0.3)',
              borderRadius: '8px',
              padding: '1rem',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--accent-cyan)' }}>
                <Lock size={16} />
                <span>HMAC Canary Signature Validated</span>
              </div>
              <div style={{ marginTop: '0.5rem', fontSize: '0.8rem', fontFamily: 'monospace', wordBreak: 'break-all', color: '#a5f3fc' }}>
                {incident.canary_hit.token}
              </div>
              <div style={{ marginTop: '0.4rem', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                Document ID: {incident.canary_hit.payload?.id} | Target Vector Store: {incident.canary_hit.payload?.st}
              </div>
            </div>
          )}

          {incident.tool_probe && (
            <div style={{
              backgroundColor: 'rgba(239, 68, 68, 0.1)',
              border: '1px solid rgba(239, 68, 68, 0.3)',
              borderRadius: '8px',
              padding: '1rem',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', fontWeight: 600, color: 'var(--danger)' }}>
                <Terminal size={16} />
                <span>Decoy Tool Invocation Intercepted</span>
              </div>
              <div style={{ marginTop: '0.5rem', fontSize: '0.85rem', fontWeight: 600 }}>
                Tool: <code style={{ color: '#fca5a5' }}>{incident.tool_probe.tool_name}</code>
              </div>
              <pre style={{
                marginTop: '0.5rem',
                padding: '0.5rem',
                backgroundColor: '#050811',
                borderRadius: '6px',
                fontSize: '0.75rem',
                color: '#e2e8f0',
                overflowX: 'auto',
              }}>
                {JSON.stringify(incident.tool_probe.arguments, null, 2)}
              </pre>
            </div>
          )}

          {/* Turn by turn trajectory playback */}
          <div>
            <h3 style={{ fontSize: '0.9rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <Clock size={16} />
              <span>Turn Trajectory History</span>
            </h3>

            {(!incident.trajectory || incident.trajectory.length === 0) ? (
              <div style={{ padding: '1rem', backgroundColor: 'var(--bg-surface)', borderRadius: '8px', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                No prior conversational turns recorded before trigger activation. Incident triggered on initial query.
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {incident.trajectory.map((step, idx) => (
                  <div key={idx} style={{
                    backgroundColor: 'var(--bg-surface)',
                    border: `1px solid ${step.canary_tripped || step.tool_tripped ? 'rgba(239, 68, 68, 0.4)' : 'var(--border-subtle)'}`,
                    borderRadius: '8px',
                    padding: '0.9rem',
                    fontSize: '0.8rem',
                  }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.4rem' }}>
                      <span style={{ fontWeight: 600, color: 'var(--accent-cyan)' }}>Turn #{idx + 1}</span>
                      <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                        {new Date(step.timestamp).toLocaleTimeString()}
                      </span>
                    </div>

                    <div style={{ color: 'var(--text-main)', marginBottom: '0.4rem' }}>
                      <strong>Prompt:</strong> {step.prompt}
                    </div>

                    {step.response && (
                      <div style={{ color: '#a5f3fc', backgroundColor: '#050811', padding: '0.5rem', borderRadius: '4px', marginTop: '0.4rem' }}>
                        <strong>Airlock Response:</strong> {step.response}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Modal Footer */}
        <div style={{
          padding: '1rem 1.5rem',
          borderTop: '1px solid var(--border-subtle)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          backgroundColor: 'var(--bg-surface)',
        }}>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button
              onClick={() => downloadExport('markdown')}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.35rem',
                padding: '0.45rem 0.85rem',
                backgroundColor: 'var(--bg-card)',
                color: 'var(--accent-cyan)',
                border: '1px solid var(--border-subtle)',
                borderRadius: '6px',
                fontSize: '0.75rem',
                fontWeight: 600,
              }}
            >
              <FileText size={14} />
              <span>Export Dossier (Markdown)</span>
            </button>

            <button
              onClick={() => downloadExport('stix')}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.35rem',
                padding: '0.45rem 0.85rem',
                backgroundColor: 'var(--bg-card)',
                color: 'var(--text-main)',
                border: '1px solid var(--border-subtle)',
                borderRadius: '6px',
                fontSize: '0.75rem',
                fontWeight: 600,
              }}
            >
              <Download size={14} />
              <span>Export STIX 2.1 (JSON)</span>
            </button>
          </div>

          <button
            onClick={onClose}
            style={{
              padding: '0.45rem 1rem',
              backgroundColor: 'var(--bg-card)',
              color: 'var(--text-main)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '6px',
              fontSize: '0.8rem',
            }}
          >
            Close Dossier
          </button>
        </div>
      </div>
    </div>
  );
};
