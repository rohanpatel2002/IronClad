import React from 'react';

interface DeploymentRiskProps {
  deploymentId: string;
  blastRadius: number;
  reversibility: number;
  timingRisk: number;
  decision: 'ALLOW' | 'WARN' | 'BLOCK';
}

export const RiskCard: React.FC<DeploymentRiskProps> = ({
  deploymentId,
  blastRadius,
  reversibility,
  timingRisk,
  decision,
}) => {
  const decisionColor = {
    'ALLOW': 'bg-green-100 text-green-800',
    'WARN': 'bg-yellow-100 text-yellow-800',
    'BLOCK': 'bg-red-100 text-red-800',
  }[decision];

  return (
    <div className="p-6 bg-white rounded-lg shadow-md border border-gray-200">
      <div className="flex justify-between items-start mb-4">
        <h3 className="text-lg font-semibold text-gray-900">{deploymentId}</h3>
        <span className={`px-3 py-1 rounded-full text-sm font-medium ${decisionColor}`}>
          {decision}
        </span>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div>
          <p className="text-sm text-gray-600">Blast Radius</p>
          <p className="text-2xl font-bold text-gray-900">{(blastRadius * 100).toFixed(0)}%</p>
        </div>
        <div>
          <p className="text-sm text-gray-600">Reversibility</p>
          <p className="text-2xl font-bold text-gray-900">{(reversibility * 100).toFixed(0)}%</p>
        </div>
        <div>
          <p className="text-sm text-gray-600">Timing Risk</p>
          <p className="text-2xl font-bold text-gray-900">{(timingRisk * 100).toFixed(0)}%</p>
        </div>
      </div>
    </div>
  );
};
