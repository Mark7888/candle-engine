import { useState, useEffect } from 'react';
import { ChevronsLeft, ChevronsRight } from 'lucide-react';
import { api } from '../../api/client';
import ProjectList from '../project/ProjectList';
import ProjectView from '../project/ProjectView';
import type { ProjectSummary } from '../../types/api';

export default function AppLayout() {
  const [projects, setProjects] = useState<ProjectSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);

  useEffect(() => {
    api
      .listProjects()
      .then((res) => {
        setProjects(res.projects);
        if (res.projects.length > 0) setSelectedId(res.projects[0].id);
      })
      .catch(() => setError('Failed to load projects.'))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="flex h-screen bg-gray-900 text-gray-100 overflow-hidden">
      {/* Sidebar */}
      <aside
        className={`flex-shrink-0 transition-all duration-300 bg-gray-850 border-r border-gray-700 flex flex-col ${
          sidebarOpen ? 'w-56' : 'w-12'
        }`}
        style={{ background: '#111827' }}
      >
        {/* Sidebar header */}
        <div className="flex items-center justify-between px-3 py-3 border-b border-gray-700 flex-shrink-0">
          {sidebarOpen && (
            <span className="text-sm font-semibold text-gray-300 truncate">Projects</span>
          )}
          <button
            onClick={() => setSidebarOpen((v) => !v)}
            className="p-1 rounded hover:bg-gray-700 text-gray-400 hover:text-gray-200 transition-colors ml-auto"
            title={sidebarOpen ? 'Collapse sidebar' : 'Expand sidebar'}
          >
            {sidebarOpen ? <ChevronsLeft size={20} /> : <ChevronsRight size={20} />}
          </button>
        </div>

        {/* Project list (only when expanded) */}
        {sidebarOpen && (
          <div className="flex-1 overflow-y-auto py-2 px-2">
            <ProjectList
              projects={projects}
              loading={loading}
              error={error}
              selectedId={selectedId}
              onSelect={setSelectedId}
            />
          </div>
        )}

        {/* Collapsed icon buttons */}
        {!sidebarOpen && (
          <div className="flex-1 flex flex-col items-center py-2 gap-1">
            {projects.map((p) => (
              <button
                key={p.id}
                onClick={() => { setSelectedId(p.id); setSidebarOpen(true); }}
                title={p.title}
                className={`w-8 h-8 rounded-md text-xs font-bold transition-colors ${
                  p.id === selectedId
                    ? 'bg-indigo-600 text-white'
                    : 'bg-gray-700 text-gray-400 hover:bg-gray-600 hover:text-gray-200'
                }`}
              >
                {p.title.charAt(0).toUpperCase()}
              </button>
            ))}
          </div>
        )}
      </aside>

      {/* Main content */}
      <main className="flex-1 flex flex-col overflow-hidden p-4 gap-3">
        {/* App header */}
        <header className="flex items-center gap-3 flex-shrink-0">
          <h1 className="text-xl font-bold text-gray-100">Candle Engine</h1>
          <span className="text-xs text-gray-500 border border-gray-700 rounded px-2 py-0.5">Chart Testing</span>
        </header>

        {/* Project view */}
        <div className="flex-1 overflow-hidden">
          {selectedId !== null ? (
            <ProjectView key={selectedId} projectId={selectedId} />
          ) : (
            <div className="flex items-center justify-center h-full text-gray-500">
              Select a project from the sidebar.
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
