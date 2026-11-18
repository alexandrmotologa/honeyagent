import React from 'react';
import { SeedResult } from '../types';
import { Database, ShieldCheck, AlertCircle } from 'lucide-react';

interface CanaryTableProps {
  canaries: SeedResult[];
}

export const CanaryTable: React.FC<CanaryTableProps> = ({ canaries }) => {
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
          <Database size={18} color="var(--accent-cyan)" />
          <h2 style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)' }}>
            Vector Store Canary Inventory
          </h2>
        </div>
        <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          {canaries.length} active honeypots
        </span>
      </div>

      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem', textAlign: 'left' }}>
          <thead>
            <tr style={{ backgroundColor: 'var(--bg-surface)', borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-muted)' }}>
              <th style={{ padding: '0.75rem 1rem', fontWeight: 600 }}>Document Title</th>
              <th style={{ padding: '0.75rem 1rem', fontWeight: 600 }}>Store Target</th>
              <th style={{ padding: '0.75rem 1rem', fontWeight: 600 }}>Category</th>
              <th style={{ padding: '0.75rem 1rem', fontWeight: 600 }}>Cryptographic Token</th>
              <th style={{ padding: '0.75rem 1rem', fontWeight: 600 }}>Status</th>
            </tr>
          </thead>
          <tbody>
            {canaries.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
                  No canaries currently seeded. Click "Seed Canaries" above to deploy lures.
                </td>
              </tr>
            ) : (
              canaries.map((c) => (
                <tr key={c.document_id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  <td style={{ padding: '0.75rem 1rem', fontWeight: 500, color: 'var(--text-main)' }}>
                    {c.title}
                  </td>
                  <td style={{ padding: '0.75rem 1rem' }}>
                    <span style={{
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      backgroundColor: 'var(--bg-surface)',
                      color: 'var(--accent-cyan)',
                      fontSize: '0.75rem',
                      fontFamily: 'monospace',
                    }}>
                      {c.store}
                    </span>
                  </td>
                  <td style={{ padding: '0.75rem 1rem', color: 'var(--text-muted)' }}>
                    {c.category}
                  </td>
                  <td style={{ padding: '0.75rem 1rem', fontFamily: 'monospace', color: '#a5f3fc', fontSize: '0.7rem' }}>
                    {c.token.length > 32 ? `${c.token.substring(0, 32)}...` : c.token}
                  </td>
                  <td style={{ padding: '0.75rem 1rem' }}>
                    <span style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '0.3rem',
                      fontSize: '0.7rem',
                      fontWeight: 600,
                      padding: '0.2rem 0.5rem',
                      borderRadius: '9999px',
                      backgroundColor: c.status === 'active' ? 'var(--success-bg)' : 'var(--danger-bg)',
                      color: c.status === 'active' ? 'var(--success)' : 'var(--danger)',
                    }}>
                      {c.status === 'active' ? <ShieldCheck size={12} /> : <AlertCircle size={12} />}
                      {c.status.toUpperCase()}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
