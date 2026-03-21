import { useState, useEffect } from 'react';
import { ChevronLeft } from 'lucide-react';
import { Spinner, Alert } from '@chakra-ui/react';
import PairTabs from './PairTabs';
import ChartView from '../chart/ChartView';
import LogsPanel from '../logs/LogsPanel';
import { api } from '../../api/client';
import type {
  ProjectDetailResponse,
  ProjectRenderResponse,
  SourcePair,
} from '../../types/api';

interface ProjectViewProps {
  projectId: number;
}

export default function ProjectView({ projectId }: ProjectViewProps) {
  const [detail, setDetail] = useState<ProjectDetailResponse | null>(null);
  const [detailLoading, setDetailLoading] = useState(true);
  const [detailError, setDetailError] = useState<string | null>(null);

  const [activePair, setActivePair] = useState<SourcePair | null>(null);
  const [activeTimeFrame, setActiveTimeFrame] = useState<string>('');

  const [renderData, setRenderData] = useState<ProjectRenderResponse | null>(null);
  const [renderLoading, setRenderLoading] = useState(false);
  const [renderError, setRenderError] = useState<string | null>(null);

  const [logsOpen, setLogsOpen] = useState(true);

  // Load project details on mount (component is remounted via key when projectId changes)
  useEffect(() => {
    let cancelled = false;
    api
      .getProject(projectId)
      .then((d) => {
        if (cancelled) return;
        setDetail(d);
        setActivePair(d.activePair);
        setActiveTimeFrame(d.timeFrame);
        setDetailLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setDetailError('Failed to load project details.');
        setDetailLoading(false);
      });
    return () => { cancelled = true; };
  }, [projectId]);

  // Load render data when project/pair/timeframe changes
  useEffect(() => {
    if (!activePair || !activeTimeFrame) return;
    let cancelled = false;
    api
      .renderProject(projectId, {
        source: activePair.sourceName,
        pair: activePair.pairId,
        timeFrame: activeTimeFrame,
      })
      .then((d) => {
        if (cancelled) return;
        setRenderData(d);
        setRenderError(null);
        setRenderLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setRenderData(null);
        setRenderError('Failed to load chart data.');
        setRenderLoading(false);
      });
    return () => { cancelled = true; };
  }, [projectId, activePair, activeTimeFrame]);

  if (detailLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="xl" />
      </div>
    );
  }

  if (detailError) {
    return (
      <Alert.Root status="error">
        <Alert.Indicator />
        <Alert.Title>{detailError}</Alert.Title>
      </Alert.Root>
    );
  }

  if (!detail) return null;

  return (
    <div className="flex flex-col gap-3 h-full">
      {/* Settings bar */}
      <div className="flex items-center gap-3 px-2 pb-3 border-b border-gray-700 flex-shrink-0 flex-wrap">
        <span className="text-sm font-semibold text-gray-300">{detail.title}</span>

        {/* Timeframe selector */}
        <div className="flex items-center gap-2">
          <label className="text-xs text-gray-500">Timeframe:</label>
          <select
            value={activeTimeFrame}
            onChange={(e) => setActiveTimeFrame(e.target.value)}
            className="px-2 py-1 text-xs bg-gray-700 border border-gray-600 rounded text-gray-200 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          >
            {detail.availableTimeFrames.map((tf) => (
              <option key={tf} value={tf}>{tf}</option>
            ))}
          </select>
        </div>

        {/* Logs toggle */}
        <button
          onClick={() => setLogsOpen((v) => !v)}
          className="ml-auto px-3 py-1 text-xs rounded bg-gray-700 hover:bg-gray-600 text-gray-300 transition-colors"
        >
          {logsOpen ? 'Hide logs' : 'Show logs'}
        </button>
      </div>

      {/* Main layout: pair tabs | chart | logs */}
      <div className="flex gap-3 flex-1 min-h-0">
        {/* Pair tabs */}
        <PairTabs
          pairs={detail.pairs}
          activePair={activePair}
          onSelect={(pair) => setActivePair(pair)}
        />

        {/* Chart area */}
        <div className="flex-1 overflow-y-auto bg-gray-800 border border-gray-700 rounded-lg p-3 min-w-0">
          <ChartView
            data={renderData}
            loading={renderLoading}
            error={renderError}
          />
        </div>

        {/* Logs panel */}
        {logsOpen && (
          <div className="w-80 flex-shrink-0 overflow-hidden">
            <LogsPanel
              logs={renderData?.logs ?? []}
              onClose={() => setLogsOpen(false)}
            />
          </div>
        )}

        {/* Collapsed logs bar */}
        {!logsOpen && (
          <button
            onClick={() => setLogsOpen(true)}
            title="Open logs panel"
            className="w-10 flex-shrink-0 bg-gray-800 border border-gray-700 rounded-lg flex items-center justify-center hover:bg-gray-700 transition-colors group"
          >
            <ChevronLeft
              size={16}
              className="text-gray-500 group-hover:text-indigo-400 transition-colors"
            />
          </button>
        )}
      </div>
    </div>
  );
}

