import React, { useEffect, useState } from 'react';
import { RiskCard } from '../components/RiskCard';
import { ThreatMap } from '../components/ThreatMap';
import { GovernanceCard } from '../components/GovernanceCard';
import { AuditLogViewer } from '../components/AuditLogViewer';

export default function Home() {
  const [deployments, setDeployments] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setDeployments([
      {
        id: 'deploy-001 (payment-api)',
        blastRadius: 0.65,
        reversibility: 0.85,
        timingRisk: 0.45,
        confidence: 0.96,
        decision: 'ALLOW',
      },
      {
        id: 'deploy-002 (auth-service)',
        blastRadius: 0.92,
        reversibility: 0.20,
        timingRisk: 0.78,
        confidence: 0.88,
        decision: 'BLOCK',
      },
    ]);
    setLoading(false);
  }, []);

  const sampleAuditLogs = [
    { id: '1', timestamp: '2026-08-28 10:00', actor: 'ci-bot', action: 'EVALUATE_DEPLOYMENT', status: 'SUCCESS' },
    { id: '2', timestamp: '2026-08-28 10:05', actor: 'alice@co.com', action: 'POLICY_UPDATE', status: 'SUCCESS' },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <h1 className="text-3xl font-bold text-gray-900">IRONCLAD Dashboard</h1>
        <p className="text-gray-600 mt-1">Autonomous Security & Risk Engine</p>
      </header>

      <main className="p-6 space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <ThreatMap maliciousIpCount={3} blockedSubnetCount={1} lastUpdated="10:00 IST" />
          <GovernanceCard status="PASSING" totalDeploys={42} allowedDeploys={38} blockedDeploys={4} />
        </div>

        <h2 className="text-xl font-bold text-gray-800">Recent Deployment Risk Evaluations</h2>
        {loading ? (
          <div className="text-center py-12">
            <p className="text-gray-500">Loading deployments...</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {deployments.map((deployment) => (
              <RiskCard
                key={deployment.id}
                deploymentId={deployment.id}
                blastRadius={deployment.blastRadius}
                reversibility={deployment.reversibility}
                timingRisk={deployment.timingRisk}
                confidence={deployment.confidence}
                decision={deployment.decision}
              />
            ))}
          </div>
        )}

        <AuditLogViewer logs={sampleAuditLogs} />
      </main>
    </div>
  );
}

