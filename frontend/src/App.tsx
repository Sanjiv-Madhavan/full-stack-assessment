import './App.css'

import '@mantine/core/styles.css';

import { Header } from './components/Header';
import Dashboard from './pages/Dashboard';

export default function App() {
  return <>
    <Header />
    <main>
      <Dashboard />
    </main>
  </>;
}

