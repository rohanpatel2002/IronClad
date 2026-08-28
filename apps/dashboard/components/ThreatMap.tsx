import React from 'react';

interface ThreatFeedProps {
  maliciousIpCount: number;
  blockedSubnetCount: number;
  lastUpdated: string;
}

export const ThreatMap: React.FC<ThreatFeedProps> = ({
  maliciousIpCount,
  blockedSubnetCount,
  lastUpdated,
}) => {
  return (
    <div className="p-6 bg-slate-900 text-white rounded-lg shadow-lg">
      <h3 className="text-lg font-bold mb-2 text-red-400">Real-Time Threat Intelligence</h3>
      <div className="grid grid-cols-2 gap-4 my-4">
        <div className="bg-slate-800 p-3 rounded">
          <p className="text-xs text-slate-400">Flagged IPs</p>
          <p className="text-xl font-bold text-red-500">{maliciousIpCount}</p>
        </div>
        <div className="bg-slate-800 p-3 rounded">
          <p className="text-xs text-slate-400">Blocked Subnets</p>
          <p className="text-xl font-bold text-orange-400">{blockedSubnetCount}</p>
        </div>
      </div>
      <p className="text-xs text-slate-400">Feed Last Sync: {lastUpdated}</p>
    </div>
  );
};
