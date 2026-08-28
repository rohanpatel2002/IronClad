import React from 'react';

interface TopologyCardProps {
  serviceName: string;
  criticality: number;
  dependencies: string[];
  dependents: string[];
}

export const TopologyCard: React.FC<TopologyCardProps> = ({
  serviceName,
  criticality,
  dependencies,
  dependents,
}) => {
  return (
    <div className="p-6 bg-white rounded-lg shadow-md border border-gray-200">
      <div className="flex justify-between items-center mb-3">
        <h4 className="font-semibold text-gray-800">{serviceName}</h4>
        <span className="text-xs px-2 py-1 bg-blue-50 text-blue-700 rounded font-mono">
          Criticality: {(criticality * 100).toFixed(0)}%
        </span>
      </div>
      <div className="text-xs text-gray-600 space-y-1">
        <p><strong>Depends On:</strong> {dependencies.join(', ') || 'None'}</p>
        <p><strong>Depended On By:</strong> {dependents.join(', ') || 'None'}</p>
      </div>
    </div>
  );
};
