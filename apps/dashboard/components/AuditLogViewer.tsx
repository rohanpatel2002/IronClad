import React from 'react';

interface AuditLog {
  id: string;
  timestamp: string;
  actor: string;
  action: string;
  status: string;
}

export const AuditLogViewer: React.FC<{ logs: AuditLog[] }> = ({ logs }) => {
  return (
    <div className="bg-slate-900 text-slate-100 p-4 rounded-lg font-mono text-xs overflow-x-auto">
      <h4 className="text-slate-400 font-bold mb-2">Audit Log Stream</h4>
      {logs.map((log) => (
        <div key={log.id} className="py-1 border-b border-slate-800 flex justify-between">
          <span className="text-slate-500">{log.timestamp}</span>
          <span className="text-cyan-400">{log.actor}</span>
          <span className="text-yellow-300">{log.action}</span>
          <span className={log.status === 'SUCCESS' ? 'text-green-400' : 'text-red-400'}>{log.status}</span>
        </div>
      ))}
    </div>
  );
};
