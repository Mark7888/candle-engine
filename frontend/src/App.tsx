import { ChakraProvider, defaultSystem } from '@chakra-ui/react';
import AppLayout from './components/layout/AppLayout';
import './index.css';

function App() {
  return (
    <ChakraProvider value={defaultSystem}>
      <AppLayout />
    </ChakraProvider>
  );
}

export default App;
