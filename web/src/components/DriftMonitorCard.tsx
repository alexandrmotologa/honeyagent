import React, { useState, useEffect } from 'react';
import { DriftReport } from '../types';
import { Compass, RefreshCw, CheckCircle2, AlertCircle } from 'lucide-react';

export const DriftMonitorCard: React.FC = () => {
  const [report, setReport] = useState<DriftReport | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchDrift = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/canaries/drift');
      const data: DriftReport = await res.json();
      setReport(data);
    } catch (err) {
      console.error('Failed to load drift report:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDrift();
  }, []);

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
          <Compass size={18} color="var(--accent-cyan)" />
          <h2 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-main)' }}>
            Canary Semantic Drift & Retrieval Ranking Monitor
          </h2>
        </div>
        <button
          onClick={fetchDrift}
          disabled={loading}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.35rem',
            padding: '0.4rem 0.75rem',
            backgroundColor: 'var(--bg-surface)',
            color: 'var(--text-main)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '6px',
            fontSize: '0.75rem',
          }}
        >
          <RefreshCw size={12} className={loading ? 'animate-pulse' : ''} />
          <span>{loading ? 'Evaluating...' : 'Run Benchmark'}</span>
        </button>
      </div>

      <div style={{ padding: '1.25rem' }}>
        {report && (
          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem' }}>
              <div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Semantic Capture Efficacy</div>
                <div style={{ fontSize: '1.5rem', fontWeight: 700, color: report.efficacy_rate >= 80 ? 'var(--success)' : 'var(--warning)' }}>
                  {report.efficacy_rate.toFixed(0)}% <span style={{ fontSize: '0.8rem', fontWeight: 500, color: 'var(--text-muted)' }}>({report.trapped_queries}/{report.total_queries} queries captured)</span>
                </div>
              </div>

              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.4rem',
                padding: '0.4rem 0.8rem',
                borderRadius: '9999px',
                backgroundColor: report.status === 'healthy' ? 'var(--success-bg)' : 'var(--warning-bg)',
                color: report.status === 'healthy' ? 'var(--success)' : 'var(--warning)',
                fontSize: '0.75rem',
                fontWeight: 600,
              }}>
                {report.status === 'healthy' ? <CheckCircle2 size={14} /> : <AlertCircle size={14} />}
                <span>{report.status.toUpperCase()}</span>
              </div>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {report.benchmarks.map((b, idx) => (
                <div
                  key={idx}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '0.65rem 0.85rem',
                    backgroundColor: 'var(--bg-surface)',
                    borderRadius: '6px',
                    fontSize: '0.75rem',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                    <span style={{
                      padding: '0.15rem 0.4rem',
                      borderRadius: '4px',
                      backgroundColor: b.is_trapped ? 'var(--success-bg)' : 'var(--danger-bg)',
                      color: b.is_trapped ? 'var(--success)' : 'var(--danger)',
                      fontWeight: 700,
                      fontSize: '0.7rem',
                    }}>
                      Rank #{b.rank}
                    </span>
                    <span style={{ color: 'var(--text-main)', fontWeight: 500 }}>
                      {b.query}
                    </span>
                  </div>

                  <span style={{ color: 'var(--text-muted)', fontFamily: 'monospace' }}>
                    score: {b.similarity_score.toFixed(3)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
