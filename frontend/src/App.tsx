import { Routes, Route } from 'react-router-dom'
import { MainLayout } from '@/components/layout/main-layout'
import { Dashboard } from '@/pages/Dashboard'
import { AgentsPage } from '@/pages/AgentsPage'
import { ArchivedAgentsPage } from '@/pages/ArchivedAgentsPage'
import { CreateAgentPage } from '@/pages/CreateAgentPage'
import { AgentDetailPage } from '@/pages/AgentDetailPage'
import { AgentVersionHistoryPage } from '@/pages/AgentVersionHistoryPage'

function App() {
  return (
    <MainLayout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/agents" element={<AgentsPage />} />
        <Route path="/agents/archived" element={<ArchivedAgentsPage />} />
        <Route path="/agents/new" element={<CreateAgentPage />} />
        <Route path="/agents/:id" element={<AgentDetailPage />} />
        <Route path="/agents/:id/versions" element={<AgentVersionHistoryPage />} />
        <Route path="/settings" element={
          <div className="space-y-6">
            <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
            <p className="text-muted-foreground">Settings page - Coming soon</p>
          </div>
        } />
      </Routes>
    </MainLayout>
  )
}

export default App