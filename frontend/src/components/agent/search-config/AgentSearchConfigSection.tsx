import { Separator } from '@/components/ui/separator'
import { AgentSlackSearchConfigCard } from './slack/AgentSlackSearchConfigCard'
import { AgentJiraSearchConfigCard } from './jira/AgentJiraSearchConfigCard'
import { AgentNotionSearchConfigCard } from './notion/AgentNotionSearchConfigCard'

interface Props {
  agentId: string
  canEdit: boolean
}

export function AgentSearchConfigSection({ agentId, canEdit }: Props) {
  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium mb-1">Search Configuration</h3>
        <p className="text-sm text-muted-foreground">
          Configure external services for the agent to search through when processing queries.
        </p>
      </div>
      
      <Separator />
      
      <div className="space-y-6">
        <AgentSlackSearchConfigCard agentId={agentId} canEdit={canEdit} />
        <AgentJiraSearchConfigCard agentId={agentId} canEdit={canEdit} />
        <AgentNotionSearchConfigCard agentId={agentId} canEdit={canEdit} />
      </div>
    </div>
  )
}