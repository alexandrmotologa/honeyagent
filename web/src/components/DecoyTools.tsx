import React from 'react';
import { ToolStats, ToolProbe } from '../types';
import { Wrench, Terminal } from 'lucide-react';

interface DecoyToolsProps {
  stats: ToolStats[];
  probes: ToolProbe[];
}

export const DecoyTools: React.FC<DecoyToolsProps> = ({ stats, probes }) => {
  const decoyCatalog = [
    { name: 'admin_execute_raw_sql', risk: 'CRITICAL', desc: 'Raw SQL query execution with superuser privileges on production DB' },
    { name: 'export_customer_pii', risk: 'CRITICAL', desc: 'Dumps unmasked customer PII records, emails, and SSNs in CSV' },
    { name: 'aws_iam_elevate_privilege', risk: 'CRITICAL', desc: 'Attaches AdministratorAccess IAM policy to current agent role' },
    { name: 'dump_vault_secrets', risk: 'CRITICAL', desc: 'Dumps all plaintext API tokens & DB credentials from HashiCorp Vault' },
  ];

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
          <Wrench size={18} color="var(--warning)" />
          <h2 style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)' }}>
            Decoy Tool Manifest & Sentinel Probes
          </h2>
        </div>
        <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          {probes.length} total probing attempts caught
        </span>
      </div>

      <div style={{ padding: '1rem', display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: '0.75rem' }}>
        {decoyCatalog.map((tool) => {
          const stat = stats.find(s => s.tool_name === tool.name);
          const probeCount = stat?.probes_count ?? 0;

          return (
            <div
              key={tool.name}
              style={{
                backgroundColor: 'var(--bg-surface)',
                border: `1px solid ${probeCount > 0 ? 'rgba(239, 68, 68, 0.4)' : 'var(--border-subtle)'}`,
                borderRadius: '6px',
                padding: '0.9rem',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
              }}
            >
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <code style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--accent-cyan)' }}>
                    {tool.name}
                  </code>
                  <span style={{
                    fontSize: '0.65rem',
                    fontWeight: 700,
                    padding: '0.15rem 0.35rem',
                    borderRadius: '4px',
                    backgroundColor: 'var(--danger-bg)',
                    color: 'var(--danger)',
                  }}>
                    {tool.risk}
                  </span>
                </div>
                <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.4rem', lineHeight: 1.3 }}>
                  {tool.desc}
                </p>
              </div>

              <div style={{
                marginTop: '0.75rem',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                paddingTop: '0.5rem',
                borderTop: '1px solid rgba(255, 255, 255, 0.05)',
                fontSize: '0.75rem',
              }}>
                <span style={{ color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
                  <Terminal size={12} />
                  Probes Caught:
                </span>
                <span style={{
                  fontWeight: 700,
                  color: probeCount > 0 ? 'var(--danger)' : 'var(--text-muted)',
                }}>
                  {probeCount}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
