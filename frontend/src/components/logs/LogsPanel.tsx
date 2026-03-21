import { useState, useMemo } from 'react';
import { X, Bug, Info, AlertTriangle, AlertCircle } from 'lucide-react';
import type { LogEntry } from '../../types/api';

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

const LEVELS: LogLevel[] = ['DEBUG', 'INFO', 'WARN', 'ERROR'];

const LEVEL_META: Record<
  LogLevel,
  { icon: React.ReactNode; badgeClass: string; textClass: string }
> = {
  DEBUG: {
    icon: <Bug size={12} />,
    badgeClass: 'bg-gray-700 text-gray-300',
    textClass: 'text-gray-400',
  },
  INFO: {
    icon: <Info size={12} />,
    badgeClass: 'bg-green-900/50 text-green-400',
    textClass: 'text-green-400',
  },
  WARN: {
    icon: <AlertTriangle size={12} />,
    badgeClass: 'bg-yellow-900/50 text-yellow-400',
    textClass: 'text-yellow-400',
  },
  ERROR: {
    icon: <AlertCircle size={12} />,
    badgeClass: 'bg-red-900/50 text-red-400',
    textClass: 'text-red-400',
  },
};

interface LogsPanelProps {
  logs: LogEntry[];
  onClose: () => void;
}

export default function LogsPanel({ logs, onClose }: LogsPanelProps) {
  const [hiddenLevels, setHiddenLevels] = useState<Set<LogLevel>>(new Set());
  const [activeIndicator, setActiveIndicator] = useState<string>('');

  const indicators = useMemo(() => {
    const set = new Set(logs.map((l) => l.indicator));
    return Array.from(set).sort();
  }, [logs]);

  const filtered = useMemo(() => {
    return logs.filter((l) => {
      if (hiddenLevels.has(l.level as LogLevel)) return false;
      if (activeIndicator && l.indicator !== activeIndicator) return false;
      return true;
    });
  }, [logs, hiddenLevels, activeIndicator]);

  function toggleLevel(level: LogLevel) {
    setHiddenLevels((prev) => {
      const next = new Set(prev);
      if (next.has(level)) next.delete(level);
      else next.add(level);
      return next;
    });
  }

  return (
    <div className="flex flex-col h-full bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700 flex-shrink-0">
        <h3 className="text-base font-semibold text-gray-100">Logs</h3>
        <button
          onClick={onClose}
          className="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-gray-200 transition-colors"
          title="Close logs panel"
        >
          <X size={18} />
        </button>
      </div>

      {/* Indicator selector */}
      <div className="px-4 py-2 border-b border-gray-700 flex-shrink-0">
        <select
          value={activeIndicator}
          onChange={(e) => setActiveIndicator(e.target.value)}
          className="w-full px-2 py-1 text-sm bg-gray-700 border border-gray-600 rounded text-gray-200 focus:outline-none focus:ring-1 focus:ring-indigo-500"
        >
          <option value="">All indicators</option>
          {indicators.map((ind) => (
            <option key={ind} value={ind}>{ind}</option>
          ))}
        </select>
      </div>

      {/* Level filter buttons */}
      <div className="flex gap-2 px-4 py-2 border-b border-gray-700 flex-shrink-0 flex-wrap">
        {LEVELS.map((level) => {
          const active = !hiddenLevels.has(level);
          const meta = LEVEL_META[level];
          return (
            <button
              key={level}
              onClick={() => toggleLevel(level)}
              className={`flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded transition-all ${
                active ? meta.badgeClass : 'bg-gray-700/40 text-gray-500 opacity-50'
              }`}
            >
              {meta.icon}
              {level}
            </button>
          );
        })}
      </div>

      {/* Log lines */}
      <div className="flex-1 overflow-y-auto px-4 py-2 space-y-0.5 font-mono text-xs min-h-0">
        {filtered.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            No logs to display.
          </div>
        ) : (
          filtered.map((log, i) => {
            const meta = LEVEL_META[log.level as LogLevel] ?? LEVEL_META.DEBUG;
            const time = new Date(log.time).toLocaleTimeString();
            return (
              <div key={i} className="flex gap-2 py-0.5">
                <span className="text-gray-600 shrink-0">[{time}]</span>
                <span className={`shrink-0 flex items-center gap-1 px-1.5 rounded text-[10px] font-medium ${meta.badgeClass}`}>
                  {meta.icon}
                  {log.level}
                </span>
                <span className={`break-all ${meta.textClass}`}>{log.message}</span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
