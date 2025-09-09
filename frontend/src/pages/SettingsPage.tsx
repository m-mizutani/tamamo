import { IntegrationsSection } from '@/components/settings/IntegrationsSection'

export function SettingsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
      
      <div className="space-y-8">
        <IntegrationsSection />
      </div>
    </div>
  )
}