import { Spinner } from '@chakra-ui/react';
import type { ProjectSummary } from '../../types/api';

interface ProjectListProps {
  projects: ProjectSummary[];
  loading: boolean;
  error: string | null;
  selectedId: number | null;
  onSelect: (id: number) => void;
}

export default function ProjectList({
  projects,
  loading,
  error,
  selectedId,
  onSelect,
}: ProjectListProps) {
  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Spinner size="md" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="px-3 py-2 text-sm text-red-400">
        Failed to load projects.
      </div>
    );
  }

  if (projects.length === 0) {
    return (
      <div className="px-3 py-4 text-sm text-gray-500 text-center">
        No projects found.
      </div>
    );
  }

  return (
    <ul className="space-y-1">
      {projects.map((project) => {
        const isActive = project.id === selectedId;
        return (
          <li key={project.id}>
            <button
              onClick={() => onSelect(project.id)}
              className={`w-full text-left px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-indigo-600 text-white'
                  : 'text-gray-300 hover:bg-gray-700 hover:text-gray-100'
              }`}
            >
              {project.title}
            </button>
          </li>
        );
      })}
    </ul>
  );
}
