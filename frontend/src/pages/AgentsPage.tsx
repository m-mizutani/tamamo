import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Plus, Users, Loader2, RefreshCw } from 'lucide-react'
import { Agent, AgentListResponse, GET_AGENTS, graphqlRequest } from '@/lib/graphql'
import { useNavigate } from 'react-router-dom'

export function AgentsPage() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [totalCount, setTotalCount] = useState(0)
  const navigate = useNavigate()

  const fetchAgents = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await graphqlRequest<{ agents: AgentListResponse }>(GET_AGENTS, {
        offset: 0,
        limit: 50
      })
      setAgents(response.agents.agents)
      setTotalCount(response.agents.totalCount)
    } catch (err) {
      console.error('Failed to fetch agents:', err)
      setError(err instanceof Error ? err.message : 'Failed to load agents')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAgents()
  }, [])

  const handleCreateAgent = () => {
    navigate('/agents/new')
  }

  const handleAgentClick = (agentId: string) => {
    navigate(`/agents/${agentId}`)
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Agents</h1>
            <p className="text-muted-foreground">
              Manage your AI agents and their configurations
            </p>
          </div>
          <Button onClick={handleCreateAgent}>
            <Plus className="mr-2 h-4 w-4" />
            New Agent
          </Button>
        </div>
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">Loading agents...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Agents</h1>
            <p className="text-muted-foreground">
              Manage your AI agents and their configurations
            </p>
          </div>
          <Button onClick={handleCreateAgent}>
            <Plus className="mr-2 h-4 w-4" />
            New Agent
          </Button>
        </div>
        <div className="text-center py-12">
          <div className="text-destructive mb-4">
            <p className="text-lg font-medium">Failed to load agents</p>
            <p className="text-sm text-muted-foreground">{error}</p>
          </div>
          <Button onClick={fetchAgents} variant="outline">
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Agents</h1>
          <p className="text-muted-foreground">
            Manage your AI agents and their configurations • {totalCount} total
          </p>
        </div>
        <Button onClick={handleCreateAgent}>
          <Plus className="mr-2 h-4 w-4" />
          New Agent
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {agents.map((agent) => (
          <Card 
            key={agent.id} 
            className="cursor-pointer hover:shadow-md transition-shadow"
            onClick={() => handleAgentClick(agent.id)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center space-x-2">
                <div className="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center">
                  <Users className="h-4 w-4 text-blue-600" />
                </div>
                <div className="flex-1">
                  <CardTitle className="text-lg">{agent.name}</CardTitle>
                  <CardDescription className="text-sm">
                    {agent.agentId} • v{agent.latest}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground line-clamp-2">
                {agent.description}
              </p>
              <div className="mt-3 text-xs text-muted-foreground">
                Created {new Date(agent.createdAt).toLocaleDateString()}
              </div>
              {agent.latestVersion && (
                <div className="mt-2 text-xs text-muted-foreground">
                  {agent.latestVersion.llmProvider} • {agent.latestVersion.llmModel}
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {agents.length === 0 && !loading && (
        <div className="text-center py-12">
          <Users className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold">No agents yet</h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Get started by creating your first AI agent.
          </p>
          <Button className="mt-4" onClick={handleCreateAgent}>
            <Plus className="mr-2 h-4 w-4" />
            Create Agent
          </Button>
        </div>
      )}
    </div>
  )
}