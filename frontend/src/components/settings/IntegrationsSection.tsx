import { JiraIntegrationCard } from './JiraIntegrationCard'
import { NotionIntegrationCard } from './NotionIntegrationCard'

export function IntegrationsSection() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Integrations</h2>
        <p className="text-muted-foreground">
          Connect external services to enhance your Tamamo experience.
        </p>
      </div>
      
      <div className="grid gap-4">
        <JiraIntegrationCard />
        <NotionIntegrationCard />
      </div>
    </div>
  )
}