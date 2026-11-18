import React from 'react';
import { Incident } from '../types';
import { Target, Shield, AlertTriangle } from 'lucide-react';

interface AtlasMatrixProps {
  incidents: Incident[];
}

export const AtlasMatrix: React.FC<AtlasMatrixProps> = ({ incidents }) => {
  const techniques = [
    {
      id: 'AML.T0054',
      name: 'LLM Prompt Injection',
      tactic: 'Execution',
      desc: 'Adversary crafts adversarial prompts to bypass guardrails or trigger unauthorized actions.',
    },
    {
      id: 'AML.T0057',
      name: 'LLM Knowledge & Prompt Extraction',
      tactic: 'Credential Access & Discovery',
      desc: 'Adversary probes internal knowledge base or attempts to dump system prompt instructions.',
    },
    {
      id: 'AML.T0024',
      name: 'Exfiltration via LLM Context',
      tactic: 'Exfiltration',
      desc: 'Adversary attempts to leak confidential vector chunks or honeypot credentials.',
    },
    {
      id: 'AML.T0016',
      name: 'Obtain Capabilities / Tool Abuse',
      tactic: 'Privilege Escalation',
      desc: 'Adversary attempts to invoke unadvertised administrative tools (SQL execution, PII export).',
    },
    {
      id: 'AML.T0040',
      name: 'ML-Enabled Cyber Reconnaissance',
      tactic: 'Reconnaissance',
      desc: 'Adversary probes schema endpoints, vector indices, and external canary listeners.',
    },
    {
      id: 'AML.T0043',
      name: 'Craft Adversarial Data',
      tactic: 'Initial Access',
      desc: 'Adversary tailors payload linguistic structures to circumvent semantic similarity filters.',
    },
  ];

  const getCountForTechnique = (techId: string) => {
    return incidents.filter(inc => inc.mitre_techniques?.includes(techId)).length;
  };

  return (
    <div style={{
      backgroundColor: 'var(--bg-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: '8px',
      overflow: 'hidden',
    }}>
      <div style={{
        padding: '1.25rem',
        borderBottom: '1px solid var(--border-subtle)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Target size={18} color="var(--accent-cyan)" />
          <h2 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-main)' }}>
            MITRE ATLAS Threat Matrix Alignment
          </h2>
        </div>
        <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          Adversarial Threat Landscape for Artificial-Intelligence Systems
        </span>
      </div>

      <div style={{ padding: '1.25rem', display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem' }}>
        {techniques.map(tech => {
          const count = getCountForTechnique(tech.id);
          const active = count > 0;

          return (
            <div
              key={tech.id}
              style={{
                backgroundColor: 'var(--bg-surface)',
                border: `1px solid ${active ? 'rgba(239, 68, 68, 0.4)' : 'var(--border-subtle)'}`,
                borderRadius: '8px',
                padding: '1rem',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
              }}
            >
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
                  <div>
                    <span style={{
                      fontSize: '0.7rem',
                      fontFamily: 'monospace',
                      fontWeight: 700,
                      color: active ? 'var(--danger)' : 'var(--accent-cyan)',
                    }}>
                      {tech.id}
                    </span>
                    <h3 style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginTop: '0.1rem' }}>
                      {tech.name}
                    </h3>
                  </div>

                  <span style={{
                    fontSize: '0.65rem',
                    padding: '0.2rem 0.4rem',
                    borderRadius: '4px',
                    backgroundColor: active ? 'var(--danger-bg)' : 'rgba(255,255,255,0.05)',
                    color: active ? 'var(--danger)' : 'var(--text-muted)',
                    fontWeight: 600,
                  }}>
                    {tech.tactic}
                  </span>
                </div>

                <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', lineHeight: 1.4 }}>
                  {tech.desc}
                </p>
              </div>

              <div style={{
                marginTop: '1rem',
                paddingTop: '0.5rem',
                borderTop: '1px solid rgba(255, 255, 255, 0.05)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                fontSize: '0.75rem',
              }}>
                <span style={{ color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
                  {active ? <AlertTriangle size={14} color="var(--danger)" /> : <Shield size={14} color="var(--success)" />}
                  Detections Caught:
                </span>
                <span style={{
                  fontWeight: 700,
                  color: active ? 'var(--danger)' : 'var(--text-muted)',
                }}>
                  {count}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
