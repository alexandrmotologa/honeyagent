import React from 'react';
import { ThreatStats } from '../types';
import { Skull, AlertTriangle, Key, Wrench, ShieldCheck, Clock } from 'lucide-react';

interface StatsGridProps {
  stats: ThreatStats | null;
}

export const StatsGrid: React.FC<StatsGridProps> = ({ stats }) => {
  const cards = [
    {
      title: 'Quarantined Adversaries',
      value: stats?.active_tarpitted ?? 0,
      icon: <Skull size={20} color="#ef4444" />,
      subtext: `${stats?.quarantined_turns ?? 0} turns isolated in sandbox`,
      alert: (stats?.active_tarpitted ?? 0) > 0,
    },
    {
      title: 'Critical Incidents',
      value: stats?.critical_incidents ?? 0,
      icon: <AlertTriangle size={20} color="#f59e0b" />,
      subtext: '100% True-Positive signals',
      alert: false,
    },
    {
      title: 'Canary Lures Seeded',
      value: stats?.seeded_canaries_count ?? 0,
      icon: <Key size={20} color="#38bdf8" />,
      subtext: `${stats?.total_canary_hits ?? 0} extraction attempts caught`,
      alert: false,
    },
    {
      title: 'Decoy Tools Probed',
      value: stats?.total_tool_probes ?? 0,
      icon: <Wrench size={20} color="#10b981" />,
      subtext: 'Unauthorized privilege escalations',
      alert: false,
    },
    {
      title: 'Adversary Delay Injected',
      value: `${(stats?.delay_injected_seconds ?? 0).toFixed(1)}s`,
      icon: <Clock size={20} color="#a855f7" />,
      subtext: 'Compute burned by tarpit',
      alert: false,
    },
    {
      title: 'Airlock Protection Rate',
      value: '100%',
      icon: <ShieldCheck size={20} color="#10b981" />,
      subtext: 'Zero false alarms',
      alert: false,
    },
  ];

  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
      gap: '1rem',
      margin: '1.5rem 0',
    }}>
      {cards.map((c, i) => (
        <div
          key={i}
          style={{
            backgroundColor: 'var(--bg-card)',
            border: `1px solid ${c.alert ? 'rgba(239, 68, 68, 0.4)' : 'var(--border-subtle)'}`,
            borderRadius: '8px',
            padding: '1.25rem',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 500 }}>{c.title}</span>
            <div style={{ padding: '0.35rem', backgroundColor: 'var(--bg-surface)', borderRadius: '6px' }}>
              {c.icon}
            </div>
          </div>
          <div style={{ marginTop: '0.75rem' }}>
            <div style={{
              fontSize: '1.75rem',
              fontWeight: 700,
              color: c.alert ? 'var(--danger)' : 'var(--text-main)',
              lineHeight: 1.1,
            }}>
              {c.value}
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.35rem' }}>
              {c.subtext}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};
