import React, { useEffect, useState } from 'react';
import { RiskCard } from '../components/RiskCard';

export default function Home() {
  const [deployments, setDeployments] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Placeholder: fetch deployments from API
    setDeployments([
      {
        id: 'deploy-001',
        blastRadius: 0.65,
        reversibility: 0.85,
        timingRisk: 0.45,
        decision: 'ALLOW',
      },
      {
        id: 'deploy-002',
        blastRadius: 0.92,
        reversibility: 0.20,
        timingRisk: 0.78,
        decision: 'BLOCK',
      },
    ]);
    setLoading(false);
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <h1 className="text-3xl font-bold text-gray-900">IRONCLAD Dashboard</h1>
        <p className="text-gray-600 mt-1">Deployment Risk Assessment</p>
      </header>

      <main className="p-6">
        {loading ? (
          <div className="text-center py-12">
            <p className="text-gray-500">Loading deployments...</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {deployments.map((deployment) => (
              <RiskCard
                key={deployment.id}
                deploymentId={deployment.id}
                blastRadius={deployment.blastRadius}
                reversibility={deployment.reversibility}
                timingRisk={deployment.timingRisk}
                decision={deployment.decision}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
