import { Routes, Route } from 'react-router-dom'
import { MainLayout } from '@/components/layout/main-layout'
import { Dashboard } from '@/pages/Dashboard'

function App() {
  return (
    <MainLayout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
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