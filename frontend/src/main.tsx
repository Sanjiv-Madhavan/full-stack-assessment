import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MantineProvider, ColorSchemeScript } from '@mantine/core';


const queryClient = new QueryClient();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ColorSchemeScript defaultColorScheme="dark" />
    <QueryClientProvider client={queryClient}>
      <MantineProvider
        defaultColorScheme="dark"
        theme={{
          primaryColor: 'green',
        }}
      >
        <App />
      </MantineProvider>
    </QueryClientProvider>
  </StrictMode>,
)
