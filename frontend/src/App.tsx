import { useEffect } from 'react';
import { Box, useColorModeValue } from '@chakra-ui/react';
import Navbar from './components/layout/Navbar';
import ProjectView from './components/layout/ProjectView';
import { useProjectStore } from './store/projectStore';
import { useUIStore } from './store/uiStore';

export default function App() {
  const loadProjects = useProjectStore((s) => s.loadProjects);
  const projects = useProjectStore((s) => s.projects);
  const projectsLoading = useProjectStore((s) => s.projectsLoading);

  const activeProjectId = useUIStore((s) => s.activeProjectId);
  const setActiveProjectId = useUIStore((s) => s.setActiveProjectId);

  const contentBg = useColorModeValue('gray.50', 'gray.900');

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  useEffect(() => {
    if (projects.length > 0 && activeProjectId === null) {
      setActiveProjectId(projects[0].id);
    }
  }, [projects, activeProjectId, setActiveProjectId]);

  return (
    <Box h="100vh" bg={contentBg} display="flex" flexDirection="column" overflow="hidden">
      <Navbar
        projects={projects}
        loading={projectsLoading}
        activeProjectId={activeProjectId}
        onSelectProject={setActiveProjectId}
      />
      <Box flex={1} overflow="hidden" px={6} pt={4} pb={4}>
        {activeProjectId !== null && (
          <ProjectView key={activeProjectId} projectId={activeProjectId} />
        )}
      </Box>
    </Box>
  );
}
