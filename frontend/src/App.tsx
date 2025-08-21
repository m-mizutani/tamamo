import { Routes, Route } from 'react-router-dom'
import { MainLayout } from '@/components/layout/main-layout'
import { Dashboard } from '@/pages/Dashboard'
import { AgentsPage } from '@/pages/AgentsPage'
import { ArchivedAgentsPage } from '@/pages/ArchivedAgentsPage'
import { CreateAgentPage } from '@/pages/CreateAgentPage'
import { AgentDetailPage } from '@/pages/AgentDetailPage'
import { AgentVersionHistoryPage } from '@/pages/AgentVersionHistoryPage'
import { LoginPage } from '@/pages/LoginPage'
import { AuthProvider, useAuth } from '@/contexts/AuthContext'
import { Toaster } from 'sonner'
import { Loader2 } from 'lucide-react'

function AppContent() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
      </div>
    );
  }

  if (!user) {
    return <LoginPage />;
  }

  return (
    <>
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
    <Toaster richColors position="top-right" />
    </>
  )
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  )
}

export default App