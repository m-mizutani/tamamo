import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { ArrowLeft, Loader2, AlertCircle } from 'lucide-react'
import { VersionHistory } from '@/components/agents/VersionHistory'
import { Agent, GET_AGENT, graphqlRequest } from '@/lib/graphql'

export function AgentVersionHistoryPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  
  const [agent, setAgent] = useState<Agent | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchAgent = async (signal?: AbortSignal) => {
    if (!id) return
    
    try {
      setLoading(true)
      setError(null)
      const response = await graphqlRequest<{ agent: Agent }>(GET_AGENT, { id }, signal)
      setAgent(response.agent)
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        return
      }
      console.error('Failed to fetch agent:', err)
      setError(err instanceof Error ? err.message : 'Failed to load agent')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const controller = new AbortController()
    fetchAgent(controller.signal)
    
    return () => {
      controller.abort()
    }
  }, [id])

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/agents')}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Agents
          </Button>
        </div>
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">Loading agent...</span>
        </div>
      </div>
    )
  }

  if (error || !agent) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/agents')}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Agents
          </Button>
        </div>
        <div className="text-center py-12">
          <div className="text-destructive mb-4">
            <AlertCircle className="h-8 w-8 mx-auto mb-2" />
            <p className="text-lg font-medium">{error || 'Agent not found'}</p>
          </div>
          <Button onClick={() => fetchAgent()} variant="outline">
            Retry
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={() => navigate(`/agents/${agent.id}`)}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to {agent.name}
          </Button>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Version History</h1>
            <p className="text-muted-foreground">
              {agent.agentId} • Current: v{agent.latest}
            </p>
          </div>
        </div>
      </div>

      <VersionHistory 
        agentUuid={agent.id} 
        currentVersion={agent.latest}
      />
    </div>
  )
}