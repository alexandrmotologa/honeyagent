import React, { useState, useEffect, useCallback } from 'react';
import { ThreatStats, Incident, SeedResult, ToolStats, ToolProbe } from './types';
import { Header } from './components/Header';
import { StatsGrid } from './components/StatsGrid';
import { ThreatFeed } from './components/ThreatFeed';
import { CanaryTable } from './components/CanaryTable';
import { DecoyTools } from './components/DecoyTools';
import { SimulatorCard } from './components/SimulatorCard';
import { TrajectoryModal } from './components/TrajectoryModal';

export const App: React.FC = () => {
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

    // Setup Server-Sent Events stream
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
