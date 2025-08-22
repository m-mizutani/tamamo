import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Plus, Users, Loader2, RefreshCw, ChevronLeft, ChevronRight, Archive, CheckSquare, Square, Undo2 } from 'lucide-react'
import { Agent, AgentListResponse, GET_AGENTS, GET_ALL_AGENTS, GET_AGENTS_BY_STATUS, ARCHIVE_AGENT, UNARCHIVE_AGENT, graphqlRequest } from '@/lib/graphql'
import { useNavigate } from 'react-router-dom'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { UserDisplayCompact } from '@/components/UserDisplay'
import { toast } from 'sonner'

const AGENTS_PER_PAGE = 18

type AgentFilter = 'active' | 'archived' | 'all'

export function AgentsPage() {
  const [agents, setAgents] = useState<Agent[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [totalCount, setTotalCount] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [filter, setFilter] = useState<AgentFilter>('active')
  const [selectedAgents, setSelectedAgents] = useState<Set<string>>(new Set())
  const [bulkOperationLoading, setBulkOperationLoading] = useState(false)
  const [showBulkArchiveDialog, setShowBulkArchiveDialog] = useState(false)
  const [showBulkUnarchiveDialog, setShowBulkUnarchiveDialog] = useState(false)
  const navigate = useNavigate()

  const fetchAgents = async (page: number = currentPage, currentFilter: AgentFilter = filter, signal?: AbortSignal) => {
    try {
      setLoading(true)
      setError(null)
      const offset = (page - 1) * AGENTS_PER_PAGE
      
      let response: { agents?: AgentListResponse; agentsByStatus?: AgentListResponse; allAgents?: AgentListResponse }
      
      if (currentFilter === 'active') {
        response = await graphqlRequest<{ agents: AgentListResponse }>(GET_AGENTS, {
          offset: offset,
          limit: AGENTS_PER_PAGE
        }, signal)
        setAgents(response.agents!.agents)
        setTotalCount(response.agents!.totalCount)
      } else if (currentFilter === 'archived') {
        response = await graphqlRequest<{ agentsByStatus: AgentListResponse }>(GET_AGENTS_BY_STATUS, {
          status: 'ARCHIVED',
          offset: offset,
          limit: AGENTS_PER_PAGE
        }, signal)
        setAgents(response.agentsByStatus!.agents)
        setTotalCount(response.agentsByStatus!.totalCount)
      } else {
        // For 'all', fetch all agents regardless of status
        response = await graphqlRequest<{ allAgents: AgentListResponse }>(GET_ALL_AGENTS, {
          offset: offset,
          limit: AGENTS_PER_PAGE
        }, signal)
        setAgents(response.allAgents!.agents)
        setTotalCount(response.allAgents!.totalCount)
      }
      
      setCurrentPage(page)
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        // Request was cancelled, don't update state
        return
      }
      console.error('Failed to fetch agents:', err)
      setError(err instanceof Error ? err.message : 'Failed to load agents')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const controller = new AbortController()
    fetchAgents(1, filter, controller.signal)

    return () => {
      controller.abort()
    }
  }, [filter])

  const handleCreateAgent = () => {
    navigate('/agents/new')
  }

  const handleAgentClick = (agentId: string) => {
    navigate(`/agents/${agentId}`)
  }

  const handleFilterChange = (newFilter: AgentFilter) => {
    setFilter(newFilter)
    setCurrentPage(1) // Reset to first page when filter changes
    setSelectedAgents(new Set()) // Clear selection when filter changes
  }

  const handleSelectAgent = (agentId: string, selected: boolean) => {
    setSelectedAgents(prev => {
      const newSet = new Set(prev)
      if (selected) {
        newSet.add(agentId)
      } else {
        newSet.delete(agentId)
      }
      return newSet
    })
  }

  const handleSelectAll = (selected: boolean) => {
    if (selected && agents) {
      setSelectedAgents(new Set(agents.map(agent => agent.id)))
    } else {
      setSelectedAgents(new Set())
    }
  }

  const handleBulkArchive = async () => {
    if (selectedAgents.size === 0) return

    const count = selectedAgents.size
    setShowBulkArchiveDialog(false)
    setBulkOperationLoading(true)
    try {
      const promises = Array.from(selectedAgents).map(agentId =>
        graphqlRequest(ARCHIVE_AGENT, { id: agentId })
      )
      
      await Promise.all(promises)
      
      // Refresh the list and clear selection
      await fetchAgents(currentPage, filter)
      setSelectedAgents(new Set())
      toast.success(`Successfully archived ${count} agent${count > 1 ? 's' : ''}`)
    } catch (err) {
      console.error('Failed to archive agents:', err)
      toast.error('Failed to archive agents', {
        description: err instanceof Error ? err.message : 'Unknown error occurred'
      })
    } finally {
      setBulkOperationLoading(false)
    }
  }

  const handleBulkUnarchive = async () => {
    if (selectedAgents.size === 0) return

    const count = selectedAgents.size
    setShowBulkUnarchiveDialog(false)
    setBulkOperationLoading(true)
    try {
      const promises = Array.from(selectedAgents).map(agentId =>
        graphqlRequest(UNARCHIVE_AGENT, { id: agentId })
      )
      
      await Promise.all(promises)
      
      // Refresh the list and clear selection
      await fetchAgents(currentPage, filter)
      setSelectedAgents(new Set())
      toast.success(`Successfully unarchived ${count} agent${count > 1 ? 's' : ''}`)
    } catch (err) {
      console.error('Failed to unarchive agents:', err)
      toast.error('Failed to unarchive agents', {
        description: err instanceof Error ? err.message : 'Unknown error occurred'
      })
    } finally {
      setBulkOperationLoading(false)
    }
  }

  // Pagination helpers
  const totalPages = Math.ceil(totalCount / AGENTS_PER_PAGE)
  const startItem = (currentPage - 1) * AGENTS_PER_PAGE + 1
  const endItem = Math.min(currentPage * AGENTS_PER_PAGE, totalCount)

  const handlePreviousPage = () => {
    if (currentPage > 1) {
      fetchAgents(currentPage - 1, filter)
    }
  }

  const handleNextPage = () => {
    if (currentPage < totalPages) {
      fetchAgents(currentPage + 1, filter)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            {filter === 'active' ? 'Active Agents' : filter === 'archived' ? 'Archived Agents' : 'All Agents'}
          </h1>
          <p className="text-muted-foreground">
            Manage your AI agents and their configurations
            {totalCount > 0 && (
              <span> • Showing {startItem}-{endItem} of {totalCount}</span>
            )}
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {!loading && (
            <>
              <Button variant="outline" onClick={() => navigate('/agents/archived')}>
                <Archive className="mr-2 h-4 w-4" />
                Archived
              </Button>
              <Button onClick={handleCreateAgent}>
                <Plus className="mr-2 h-4 w-4" />
                New Agent
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Filter Tabs and Bulk Actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-1 bg-muted p-1 rounded-lg w-fit">
          <Button
            variant={filter === 'active' ? 'default' : 'ghost'}
            size="sm"
            onClick={() => handleFilterChange('active')}
            className="h-8"
          >
            Active
          </Button>
          <Button
            variant={filter === 'archived' ? 'default' : 'ghost'}
            size="sm"
            onClick={() => handleFilterChange('archived')}
            className="h-8"
          >
            <Archive className="mr-1 h-3 w-3" />
            Archived
          </Button>
          <Button
            variant={filter === 'all' ? 'default' : 'ghost'}
            size="sm"
            onClick={() => handleFilterChange('all')}
            className="h-8"
          >
            All
          </Button>
        </div>

        {/* Bulk Actions */}
        {selectedAgents.size > 0 && (
          <div className="flex items-center space-x-2">
            <span className="text-sm text-muted-foreground">
              {selectedAgents.size} selected
            </span>
            {(filter === 'active' || filter === 'all') && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowBulkArchiveDialog(true)}
                disabled={bulkOperationLoading}
                className="h-8"
              >
                {bulkOperationLoading ? (
                  <Loader2 className="mr-1 h-3 w-3 animate-spin" />
                ) : (
                  <Archive className="mr-1 h-3 w-3" />
                )}
                Archive Selected
              </Button>
            )}
            {(filter === 'archived' || filter === 'all') && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowBulkUnarchiveDialog(true)}
                disabled={bulkOperationLoading}
                className="h-8"
              >
                {bulkOperationLoading ? (
                  <Loader2 className="mr-1 h-3 w-3 animate-spin" />
                ) : (
                  <Undo2 className="mr-1 h-3 w-3" />
                )}
                Unarchive Selected
              </Button>
            )}
          </div>
        )}
      </div>

      {/* Main Content Area */}
      {agents === null || loading ? (
        /* Loading State */
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">Loading agents...</span>
        </div>
      ) : error ? (
        /* Error State */
        <div className="text-center py-12">
          <div className="text-destructive mb-4">
            <p className="text-lg font-medium">Failed to load agents</p>
            <p className="text-sm text-muted-foreground">{error}</p>
          </div>
          <Button onClick={() => fetchAgents(currentPage, filter)} variant="outline">
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </div>
      ) : agents.length === 0 ? (
        /* No Agents State */
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
      ) : (
        /* Agents List State */
        <>
          {/* Select All */}
          {agents.length > 0 && (
            <div className="flex items-center space-x-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleSelectAll(!selectedAgents.size || selectedAgents.size < agents.length)}
                className="h-8"
              >
                {selectedAgents.size === agents.length ? (
                  <CheckSquare className="mr-1 h-3 w-3" />
                ) : (
                  <Square className="mr-1 h-3 w-3" />
                )}
                {selectedAgents.size === agents.length ? 'Deselect All' : 'Select All'}
              </Button>
            </div>
          )}

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {agents.map((agent) => (
              <Card 
            key={agent.id} 
            className={`relative hover:shadow-md transition-shadow ${
              selectedAgents.has(agent.id) ? 'ring-2 ring-blue-500' : ''
            }`}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center space-x-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    handleSelectAgent(agent.id, !selectedAgents.has(agent.id))
                  }}
                  className="h-8 w-8 p-0"
                >
                  {selectedAgents.has(agent.id) ? (
                    <CheckSquare className="h-4 w-4 text-blue-600" />
                  ) : (
                    <Square className="h-4 w-4" />
                  )}
                </Button>
                <div 
                  className="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center cursor-pointer"
                  onClick={() => handleAgentClick(agent.id)}
                >
                  <Users className="h-4 w-4 text-blue-600" />
                </div>
                <div 
                  className="flex-1 cursor-pointer"
                  onClick={() => handleAgentClick(agent.id)}
                >
                  <div className="flex items-center space-x-2">
                    <CardTitle className="text-lg">{agent.name}</CardTitle>
                    <Badge variant={agent.status === 'ACTIVE' ? 'default' : 'secondary'} className="text-xs">
                      {agent.status === 'ACTIVE' ? 'Active' : 'Archived'}
                    </Badge>
                  </div>
                  <CardDescription className="text-sm">
                    {agent.agentId} • v{agent.latest}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent 
              className="cursor-pointer"
              onClick={() => handleAgentClick(agent.id)}
            >
              <p className="text-sm text-muted-foreground line-clamp-2">
                {agent.description}
              </p>
              <div className="mt-3 flex items-center justify-between">
                <div className="text-xs text-muted-foreground">
                  Created {new Date(agent.createdAt).toLocaleDateString()}
                </div>
                <UserDisplayCompact user={agent.author} size={16} />
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
                Showing {startItem}-{endItem} of {totalCount} agents
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
        </>
      )}

      {/* Bulk Archive Confirmation Dialog */}
      <ConfirmDialog
        open={showBulkArchiveDialog}
        onOpenChange={setShowBulkArchiveDialog}
        title="Archive Selected Agents"
        description={`Are you sure you want to archive ${selectedAgents.size} agent${selectedAgents.size > 1 ? 's' : ''}? Archived agents cannot be used in Slack conversations but can be restored later.`}
        confirmText="Archive"
        confirmVariant="destructive"
        onConfirm={handleBulkArchive}
      />

      {/* Bulk Unarchive Confirmation Dialog */}
      <ConfirmDialog
        open={showBulkUnarchiveDialog}
        onOpenChange={setShowBulkUnarchiveDialog}
        title="Unarchive Selected Agents"
        description={`Are you sure you want to unarchive ${selectedAgents.size} agent${selectedAgents.size > 1 ? 's' : ''}? This will make them available for use in Slack conversations again.`}
        confirmText="Unarchive"
        confirmVariant="default"
        onConfirm={handleBulkUnarchive}
      />
    </div>
  )
}