import React, { useState, useEffect, useCallback } from 'react';
import { ThreatStats, Incident, SeedResult, ToolStats, ToolProbe } from './types';
import { Header } from './components/Header';
import { StatsGrid } from './components/StatsGrid';
import { ThreatFeed } from './components/ThreatFeed';
import { CanaryTable } from './components/CanaryTable';
import { DecoyTools } from './components/DecoyTools';
import { SimulatorCard } from './components/SimulatorCard';
import { TrajectoryModal } from './components/TrajectoryModal';
import { AirlockPlayground } from './components/AirlockPlayground';
import { AtlasMatrix } from './components/AtlasMatrix';
import { DriftMonitorCard } from './components/DriftMonitorCard';
import { Shield, Sparkles, Target } from 'lucide-react';

type Tab = 'overview' | 'playground' | 'atlas';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<Tab>('overview');
  const [stats, setStats] = useState<ThreatStats | null>(null);
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [canaries, setCanaries] = useState<SeedResult[]>([]);
  const [toolStats, setToolStats] = useState<ToolStats[]>([]);
  const [toolProbes, setToolProbes] = useState<ToolProbe[]>([]);
  const [connected, setConnected] = useState<boolean>(false);
  const [seeding, setSeeding] = useState<boolean>(false);
  const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null);

  const fetchData = useCallback(async () => {
    try {
      const [statsRes, threatsRes, canariesRes, decoysRes] = await Promise.all([
        fetch('/api/stats').then(r => r.json()),
        fetch('/api/threats').then(r => r.json()),
        fetch('/api/canaries').then(r => r.json()),
        fetch('/api/decoys').then(r => r.json()),
      ]);

      setStats(statsRes);
      setIncidents(threatsRes);
      setCanaries(canariesRes);
      if (decoysRes) {
        setToolStats(decoysRes.stats || []);
        setToolProbes(decoysRes.probes || []);
      }
    } catch (err) {
      console.error('Failed to sync honeyagent telemetry:', err);
    }
  }, []);

  const handleSeed = async () => {
    setSeeding(true);
    try {
      await fetch('/api/canaries/seed', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ store: 'memory' }),
      });
      await fetchData();
    } catch (err) {
      console.error('Failed to seed canaries:', err);
    } finally {
      setSeeding(false);
    }
  };

  useEffect(() => {
    fetchData();

    const eventSource = new EventSource('/api/events');

    eventSource.onopen = () => {
      setConnected(true);
    };

    eventSource.addEventListener('incident', (e) => {
      try {
        const newInc: Incident = JSON.parse(e.data);
        setIncidents(prev => [newInc, ...prev]);
        fetchData();
      } catch (err) {
        console.error('Error parsing incident event:', err);
      }
    });

    eventSource.onerror = () => {
      setConnected(false);
    };

    return () => {
      eventSource.close();
    };
  }, [fetchData]);

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header
        connected={connected}
        onRefresh={fetchData}
        onSeed={handleSeed}
        seeding={seeding}
      />

      <main style={{ flex: 1, padding: '1.5rem 2rem', maxWidth: '1440px', margin: '0 auto', width: '100%' }}>
        <StatsGrid stats={stats} />

        {/* Navigation Tabs */}
        <div style={{
          display: 'flex',
          gap: '0.5rem',
          borderBottom: '1px solid var(--border-subtle)',
          marginBottom: '1.5rem',
          paddingBottom: '0.5rem',
        }}>
          <button
            onClick={() => setActiveTab('overview')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              padding: '0.5rem 1rem',
              backgroundColor: activeTab === 'overview' ? 'var(--bg-surface)' : 'transparent',
              color: activeTab === 'overview' ? 'var(--accent-cyan)' : 'var(--text-muted)',
              border: `1px solid ${activeTab === 'overview' ? 'var(--border-subtle)' : 'transparent'}`,
              borderRadius: '6px',
              fontSize: '0.85rem',
              fontWeight: 600,
            }}
          >
            <Shield size={16} />
            <span>Overview & Incident Feed</span>
          </button>

          <button
            onClick={() => setActiveTab('playground')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              padding: '0.5rem 1rem',
              backgroundColor: activeTab === 'playground' ? 'var(--bg-surface)' : 'transparent',
              color: activeTab === 'playground' ? 'var(--accent-cyan)' : 'var(--text-muted)',
              border: `1px solid ${activeTab === 'playground' ? 'var(--border-subtle)' : 'transparent'}`,
              borderRadius: '6px',
              fontSize: '0.85rem',
              fontWeight: 600,
            }}
          >
            <Sparkles size={16} />
            <span>Interactive Airlock Playground</span>
          </button>

          <button
            onClick={() => setActiveTab('atlas')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              padding: '0.5rem 1rem',
              backgroundColor: activeTab === 'atlas' ? 'var(--bg-surface)' : 'transparent',
              color: activeTab === 'atlas' ? 'var(--accent-cyan)' : 'var(--text-muted)',
              border: `1px solid ${activeTab === 'atlas' ? 'var(--border-subtle)' : 'transparent'}`,
              borderRadius: '6px',
              fontSize: '0.85rem',
              fontWeight: 600,
            }}
          >
            <Target size={16} />
            <span>MITRE ATLAS & Semantic Drift</span>
          </button>
        </div>

        {/* Tab 1: Overview & Incidents */}
        {activeTab === 'overview' && (
          <div>
            <div style={{ marginBottom: '1.5rem' }}>
              <SimulatorCard onSimulationComplete={fetchData} />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: '1.5rem', marginBottom: '1.5rem' }}>
              <ThreatFeed
                incidents={incidents}
                onSelectIncident={setSelectedIncident}
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(480px, 1fr))', gap: '1.5rem' }}>
              <CanaryTable canaries={canaries} />
              <DecoyTools stats={toolStats} probes={toolProbes} />
            </div>
          </div>
        )}

        {/* Tab 2: Interactive Airlock Playground */}
        {activeTab === 'playground' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            <AirlockPlayground onIncidentTriggered={fetchData} />
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(480px, 1fr))', gap: '1.5rem' }}>
              <CanaryTable canaries={canaries} />
              <ThreatFeed
                incidents={incidents}
                onSelectIncident={setSelectedIncident}
              />
            </div>
          </div>
        )}

        {/* Tab 3: MITRE ATLAS & Semantic Drift */}
        {activeTab === 'atlas' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            <AtlasMatrix incidents={incidents} />
            <DriftMonitorCard />
          </div>
        )}
      </main>

      <footer style={{
        padding: '1.25rem 2rem',
        borderTop: '1px solid var(--border-subtle)',
        fontSize: '0.75rem',
        color: 'var(--text-muted)',
        textAlign: 'center',
        backgroundColor: 'var(--bg-card)',
      }}>
        honeyagent v1.0.0 &bull; Enterprise Autonomous AI Deception Airlock &bull; Zero False Positives
      </footer>

      {selectedIncident && (
        <TrajectoryModal
          incident={selectedIncident}
          onClose={() => setSelectedIncident(null)}
        />
      )}
    </div>
  );
};

export default App;
