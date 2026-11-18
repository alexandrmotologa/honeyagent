import React from 'react';
import { ShieldAlert, Activity, RefreshCw, Zap } from 'lucide-react';

interface HeaderProps {
  connected: boolean;
  onRefresh: () => void;
  onSeed: () => void;
  seeding: boolean;
}

export const Header: React.FC<HeaderProps> = ({ connected, onRefresh, onSeed, seeding }) => {
  return (
    <header style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '1.25rem 2rem',
      backgroundColor: 'var(--bg-card)',
      borderBottom: '1px solid var(--border-subtle)',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        <div style={{
          backgroundColor: '#0284c7',
          padding: '0.5rem',
          borderRadius: '8px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center'
        }}>
          <ShieldAlert size={24} color="#ffffff" />
        </div>
        <div>
          <h1 style={{ fontSize: '1.25rem', fontWeight: 700, letterSpacing: '-0.02em', color: 'var(--text-main)' }}>
            honeyagent <span style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--accent-cyan)', marginLeft: '0.25rem' }}>STUDIO</span>
          </h1>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Autonomous AI Threat Deception Mesh & Canary Airlock
          </p>
        </div>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.4rem',
          fontSize: '0.75rem',
          padding: '0.35rem 0.75rem',
          borderRadius: '9999px',
          backgroundColor: connected ? 'var(--success-bg)' : 'var(--warning-bg)',
          color: connected ? 'var(--success)' : 'var(--warning)',
          border: `1px solid ${connected ? 'rgba(16,185,129,0.3)' : 'rgba(245,158,11,0.3)'}`,
        }}>
          <Activity size={14} className={connected ? '' : 'animate-pulse-red'} />
          <span>{connected ? 'SSE Mesh Active' : 'Connecting...'}</span>
        </div>

        <button
          onClick={onSeed}
          disabled={seeding}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.4rem',
            padding: '0.45rem 0.9rem',
            backgroundColor: '#0284c7',
            color: '#ffffff',
            borderRadius: '6px',
            fontSize: '0.8rem',
            fontWeight: 600,
          }}
        >
          <Zap size={14} />
          <span>{seeding ? 'Seeding...' : 'Seed Canaries'}</span>
        </button>

        <button
          onClick={onRefresh}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.4rem',
            padding: '0.45rem 0.75rem',
            backgroundColor: 'var(--bg-surface)',
            color: 'var(--text-main)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '6px',
            fontSize: '0.8rem',
          }}
        >
          <RefreshCw size={14} />
          <span>Sync</span>
        </button>
      </div>
    </header>
  );
};
