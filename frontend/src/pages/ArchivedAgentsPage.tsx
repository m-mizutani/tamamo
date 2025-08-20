import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Archive, Loader2, RefreshCw, ChevronLeft, ChevronRight, Undo2 } from 'lucide-react'
import { Agent, AgentListResponse, GET_AGENTS_BY_STATUS, UNARCHIVE_AGENT, graphqlRequest } from '@/lib/graphql'
import { useNavigate } from 'react-router-dom'

const AGENTS_PER_PAGE = 18

export function ArchivedAgentsPage() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [totalCount, setTotalCount] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [unarchivingAgents, setUnarchivingAgents] = useState<Set<string>>(new Set())
  const navigate = useNavigate()

  const fetchArchivedAgents = async (page: number = currentPage, signal?: AbortSignal) => {
    try {
      setLoading(true)
      setError(null)
      const offset = (page - 1) * AGENTS_PER_PAGE
      const response = await graphqlRequest<{ agentsByStatus: AgentListResponse }>(GET_AGENTS_BY_STATUS, {
        status: 'ARCHIVED',
        offset: offset,
        limit: AGENTS_PER_PAGE
      }, signal)
      setAgents(response.agentsByStatus.agents)
      setTotalCount(response.agentsByStatus.totalCount)
      setCurrentPage(page)
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        // Request was cancelled, don't update state
        return
      }
      console.error('Failed to fetch archived agents:', err)
      setError(err instanceof Error ? err.message : 'Failed to load archived agents')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const controller = new AbortController()
    fetchArchivedAgents(1, controller.signal)

    return () => {
      controller.abort()
    }
  }, [])

  const handleAgentClick = (agentId: string) => {
    navigate(`/agents/${agentId}`)
  }

  const handleUnarchiveAgent = async (agentId: string, event: React.MouseEvent) => {
    event.stopPropagation() // Prevent card click navigation

    setUnarchivingAgents(prev => new Set(prev).add(agentId))
    
    try {
      await graphqlRequest(UNARCHIVE_AGENT, { id: agentId })
      
      // Refresh the list to remove the unarchived agent
      await fetchArchivedAgents(currentPage)
    } catch (err) {
      console.error('Failed to unarchive agent:', err)
      // For now, just log the error. In a real app, you'd want to show a proper error message
      alert('Failed to unarchive agent: ' + (err instanceof Error ? err.message : 'Unknown error'))
    } finally {
      setUnarchivingAgents(prev => {
        const newSet = new Set(prev)
        newSet.delete(agentId)
        return newSet
      })
    }
  }

  // Pagination helpers
  const totalPages = Math.ceil(totalCount / AGENTS_PER_PAGE)
  const startItem = (currentPage - 1) * AGENTS_PER_PAGE + 1
  const endItem = Math.min(currentPage * AGENTS_PER_PAGE, totalCount)

  const handlePreviousPage = () => {
    if (currentPage > 1) {
      fetchArchivedAgents(currentPage - 1)
    }
  }

  const handleNextPage = () => {
    if (currentPage < totalPages) {
      fetchArchivedAgents(currentPage + 1)
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Archived Agents</h1>
            <p className="text-muted-foreground">
              View and manage your archived AI agents
            </p>
          </div>
          <Button variant="outline" onClick={() => navigate('/agents')}>
            View Active Agents
          </Button>
        </div>
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">Loading archived agents...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Archived Agents</h1>
            <p className="text-muted-foreground">
              View and manage your archived AI agents
            </p>
          </div>
          <Button variant="outline" onClick={() => navigate('/agents')}>
            View Active Agents
          </Button>
        </div>
        <div className="text-center py-12">
          <div className="text-destructive mb-4">
            <p className="text-lg font-medium">Failed to load archived agents</p>
            <p className="text-sm text-muted-foreground">{error}</p>
          </div>
          <Button onClick={() => fetchArchivedAgents(currentPage)} variant="outline">
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
          <h1 className="text-3xl font-bold tracking-tight">Archived Agents</h1>
          <p className="text-muted-foreground">
            View and manage your archived AI agents
            {totalCount > 0 && (
              <span> • Showing {startItem}-{endItem} of {totalCount}</span>
            )}
          </p>
        </div>
        <Button variant="outline" onClick={() => navigate('/agents')}>
          View Active Agents
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {agents.map((agent) => (
          <Card 
            key={agent.id} 
            className="cursor-pointer hover:shadow-md transition-shadow border-muted-foreground/20 group"
            onClick={() => handleAgentClick(agent.id)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center space-x-2">
                <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center">
                  <Archive className="h-4 w-4 text-muted-foreground" />
                </div>
                <div className="flex-1">
                  <div className="flex items-center space-x-2">
                    <CardTitle className="text-lg text-muted-foreground">{agent.name}</CardTitle>
                    <Badge variant="secondary" className="text-xs">
                      Archived
                    </Badge>
                  </div>
                  <CardDescription className="text-sm">
                    {agent.agentId} • v{agent.latest}
                  </CardDescription>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => handleUnarchiveAgent(agent.id, e)}
                  disabled={unarchivingAgents.has(agent.id)}
                  className="opacity-0 group-hover:opacity-100 transition-opacity"
                  title="Unarchive agent"
                >
                  {unarchivingAgents.has(agent.id) ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Undo2 className="h-4 w-4" />
                  )}
                </Button>
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

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            Showing {startItem}-{endItem} of {totalCount} archived agents
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handlePreviousPage}
              disabled={currentPage <= 1}
            >
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <div className="flex items-center space-x-1">
              <span className="text-sm font-medium">
                Page {currentPage} of {totalPages}
              </span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleNextPage}
              disabled={currentPage >= totalPages}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      {agents.length === 0 && !loading && (
        <div className="text-center py-12">
          <Archive className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold">No archived agents</h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Agents you archive will appear here. You can unarchive them at any time.
          </p>
          <Button className="mt-4" variant="outline" onClick={() => navigate('/agents')}>
            View Active Agents
          </Button>
        </div>
      )}
    </div>
  )
}