import React from 'react';

interface GovernanceCardProps {
  status: 'PASSING' | 'REVIEW_REQUIRED';
  totalDeploys: number;
  allowedDeploys: number;
  blockedDeploys: number;
}

export const GovernanceCard: React.FC<GovernanceCardProps> = ({
  status,
  totalDeploys,
  allowedDeploys,
  blockedDeploys,
}) => {
  return (
    <div className="p-6 bg-white rounded-lg shadow-md border border-gray-200">
      <div className="flex justify-between items-center mb-3">
        <h4 className="font-semibold text-gray-900">SOC2 & ISO27001 Compliance</h4>
        <span className={`px-2.5 py-0.5 rounded text-xs font-bold ${status === 'PASSING' ? 'bg-green-100 text-green-800' : 'bg-amber-100 text-amber-800'}`}>
          {status}
        </span>
      </div>
      <div className="text-sm text-gray-600 space-y-1">
        <p>Total Evaluated: {totalDeploys}</p>
        <p className="text-green-600">Allowed: {allowedDeploys}</p>
        <p className="text-red-600">Blocked: {blockedDeploys}</p>
      </div>
    </div>
  );
};
