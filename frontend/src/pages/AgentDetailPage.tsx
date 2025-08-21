import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Separator } from '@/components/ui/separator'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { 
  ArrowLeft, 
  Save, 
  Edit3, 
  Archive, 
  ArchiveRestore,
  Loader2, 
  AlertCircle, 
  CheckCircle,
  History,
  Clock,
  User,
  Settings
} from 'lucide-react'
import { 
  Agent,
  GET_AGENT, 
  UPDATE_AGENT,
  CHECK_AGENT_ID_AVAILABILITY,
  ARCHIVE_AGENT,
  UNARCHIVE_AGENT,
  graphqlRequest,
  AgentIdAvailability,
  UpdateAgentInput
} from '@/lib/graphql'
import { Badge } from '@/components/ui/badge'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { CreateVersionDialog } from '@/components/agents/CreateVersionDialog'

const LLM_PROVIDERS = [
  { value: 'OPENAI', label: 'OpenAI' },
  { value: 'CLAUDE', label: 'Claude' },
  { value: 'GEMINI', label: 'Gemini' },
] as const

const COMMON_MODELS = {
  OPENAI: ['gpt-4o', 'gpt-4o-mini', 'gpt-4', 'gpt-3.5-turbo'],
  CLAUDE: ['claude-3-5-sonnet-20241022', 'claude-3-5-haiku-20241022', 'claude-3-opus-20240229'],
  GEMINI: ['gemini-2.0-flash', 'gemini-1.5-pro', 'gemini-1.5-flash'],
} as const


export function AgentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  
  const [agent, setAgent] = useState<Agent | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false)
  const [showCreateVersionDialog, setShowCreateVersionDialog] = useState(false)
  const [showArchiveDialog, setShowArchiveDialog] = useState(false)
  const [showUnarchiveDialog, setShowUnarchiveDialog] = useState(false)
  
  // Edit form state
  const [editForm, setEditForm] = useState({
    agentId: '',
    name: '',
    description: '',
    systemPrompt: '',
    llmProvider: 'OPENAI' as 'OPENAI' | 'CLAUDE' | 'GEMINI',
    llmModel: ''
  })
  
  const [agentIdStatus, setAgentIdStatus] = useState<{
    checking: boolean
    availability: AgentIdAvailability | null
  }>({ checking: false, availability: null })

  const fetchAgent = async (signal?: AbortSignal) => {
    if (!id) return
    
    try {
      setLoading(true)
      setError(null)
      const response = await graphqlRequest<{ agent: Agent }>(GET_AGENT, { id }, signal)
      setAgent(response.agent)
      setEditForm({
        agentId: response.agent.agentId,
        name: response.agent.name,
        description: response.agent.description,
        systemPrompt: response.agent.latestVersion?.systemPrompt || '',
        llmProvider: (response.agent.latestVersion?.llmProvider || 'OPENAI') as 'OPENAI' | 'CLAUDE' | 'GEMINI',
        llmModel: response.agent.latestVersion?.llmModel || ''
      })
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

  const checkAgentIdAvailability = async (agentId: string) => {
    if (!agentId || agentId === agent?.agentId) {
      setAgentIdStatus({ checking: false, availability: null })
      return
    }

    try {
      setAgentIdStatus({ checking: true, availability: null })
      const response = await graphqlRequest<{ checkAgentIdAvailability: AgentIdAvailability }>(
        CHECK_AGENT_ID_AVAILABILITY,
        { agentId }
      )
      setAgentIdStatus({ 
        checking: false, 
        availability: response.checkAgentIdAvailability 
      })
    } catch (err) {
      console.error('Failed to check agent ID availability:', err)
      setAgentIdStatus({ checking: false, availability: null })
    }
  }

  const handleEdit = () => {
    setIsEditing(true)
    setAgentIdStatus({ checking: false, availability: null })
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
    if (agent) {
      setEditForm({
        agentId: agent.agentId,
        name: agent.name,
        description: agent.description,
        systemPrompt: agent.latestVersion?.systemPrompt || '',
        llmProvider: (agent.latestVersion?.llmProvider || 'OPENAI') as 'OPENAI' | 'CLAUDE' | 'GEMINI',
        llmModel: agent.latestVersion?.llmModel || ''
      })
    }
    setAgentIdStatus({ checking: false, availability: null })
  }

  const handleSave = async () => {
    if (!agent) return
    
    try {
      setSaving(true)
      setError(null)
      
      const input: UpdateAgentInput = {}
      if (editForm.agentId !== agent.agentId) input.agentId = editForm.agentId
      if (editForm.name !== agent.name) input.name = editForm.name
      if (editForm.description !== agent.description) input.description = editForm.description
      
      // Check for version-related changes
      if (editForm.systemPrompt !== (agent.latestVersion?.systemPrompt || '')) {
        input.systemPrompt = editForm.systemPrompt
      }
      if (editForm.llmProvider !== (agent.latestVersion?.llmProvider || 'OPENAI')) {
        input.llmProvider = editForm.llmProvider
      }
      if (editForm.llmModel !== (agent.latestVersion?.llmModel || '')) {
        input.llmModel = editForm.llmModel
      }
      
      // Only update if there are changes
      if (Object.keys(input).length > 0) {
        const response = await graphqlRequest<{ updateAgent: Agent }>(UPDATE_AGENT, {
          id: agent.id,
          input
        })
        setAgent(response.updateAgent)
        
        // Refresh agent data to get updated version info
        await fetchAgent()
      }
      
      setIsEditing(false)
    } catch (err) {
      console.error('Failed to update agent:', err)
      setError(err instanceof Error ? err.message : 'Failed to update agent')
    } finally {
      setSaving(false)
    }
  }

  const handleAgentIdChange = (value: string) => {
    setEditForm(prev => ({ ...prev, agentId: value }))
    
    // Debounce the availability check
    const timeoutId = setTimeout(() => {
      checkAgentIdAvailability(value)
    }, 500)
    
    return () => clearTimeout(timeoutId)
  }

  const handleProviderChange = (provider: 'OPENAI' | 'CLAUDE' | 'GEMINI') => {
    setEditForm(prev => ({
      ...prev,
      llmProvider: provider,
      llmModel: COMMON_MODELS[provider][0] // Set first model as default
    }))
  }

  const handleArchive = async () => {
    if (!agent) return
    
    try {
      setIsUpdatingStatus(true)
      setError(null)
      
      const response = await graphqlRequest<{ archiveAgent: Agent }>(ARCHIVE_AGENT, {
        id: agent.id
      })
      
      setAgent(response.archiveAgent)
    } catch (err) {
      console.error('Failed to archive agent:', err)
      setError(err instanceof Error ? err.message : 'Failed to archive agent')
    } finally {
      setIsUpdatingStatus(false)
    }
  }

  const handleUnarchive = async () => {
    if (!agent) return
    
    try {
      setIsUpdatingStatus(true)
      setError(null)
      
      const response = await graphqlRequest<{ unarchiveAgent: Agent }>(UNARCHIVE_AGENT, {
        id: agent.id
      })
      
      setAgent(response.unarchiveAgent)
    } catch (err) {
      console.error('Failed to unarchive agent:', err)
      setError(err instanceof Error ? err.message : 'Failed to unarchive agent')
    } finally {
      setIsUpdatingStatus(false)
    }
  }

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

  if (error) {
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
            <p className="text-lg font-medium">Failed to load agent</p>
            <p className="text-sm text-muted-foreground">{error}</p>
          </div>
          <Button onClick={() => fetchAgent()} variant="outline">
            Retry
          </Button>
        </div>
      </div>
    )
  }

  if (!agent) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/agents')}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Agents
          </Button>
        </div>
        <div className="text-center py-12">
          <p className="text-lg font-medium">Agent not found</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/agents')}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Agents
          </Button>
          <div>
            <div className="flex items-center space-x-3">
              <h1 className="text-3xl font-bold tracking-tight">{agent.name}</h1>
              <Badge variant={agent.status === 'ACTIVE' ? 'default' : 'secondary'}>
                {agent.status === 'ACTIVE' ? 'Active' : 'Archived'}
              </Badge>
            </div>
            <p className="text-muted-foreground">
              {agent.agentId} • v{agent.latest}
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          {!isEditing ? (
            <Button onClick={handleEdit}>
              <Edit3 className="mr-2 h-4 w-4" />
              Edit
            </Button>
          ) : (
            <>
              <Button variant="outline" onClick={handleCancelEdit}>
                Cancel
              </Button>
              <Button 
                onClick={handleSave} 
                disabled={saving || (editForm.agentId !== agent.agentId && !agentIdStatus.availability?.available)}
              >
                {saving ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Save className="mr-2 h-4 w-4" />
                )}
                Save Changes
              </Button>
            </>
          )}
        </div>
      </div>

      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-destructive">
              <AlertCircle className="h-4 w-4" />
              <span className="font-medium">Error:</span>
              <span>{error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Settings className="h-5 w-5" />
                <span>Basic Information</span>
              </CardTitle>
              <CardDescription>
                Agent identity and configuration
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="agentId">Agent ID</Label>
                {isEditing ? (
                  <div className="relative">
                    <Input
                      id="agentId"
                      value={editForm.agentId}
                      onChange={(e) => handleAgentIdChange(e.target.value)}
                      className={
                        agentIdStatus.availability?.available === false 
                          ? 'border-destructive' 
                          : agentIdStatus.availability?.available === true
                          ? 'border-green-500'
                          : ''
                      }
                    />
                    {agentIdStatus.checking && (
                      <div className="absolute right-3 top-1/2 -translate-y-1/2">
                        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                      </div>
                    )}
                    {agentIdStatus.availability?.available === true && (
                      <div className="absolute right-3 top-1/2 -translate-y-1/2">
                        <CheckCircle className="h-4 w-4 text-green-500" />
                      </div>
                    )}
                    {agentIdStatus.availability?.available === false && (
                      <div className="absolute right-3 top-1/2 -translate-y-1/2">
                        <AlertCircle className="h-4 w-4 text-destructive" />
                      </div>
                    )}
                  </div>
                ) : (
                  <p className="px-3 py-2 bg-muted rounded-md">{agent.agentId}</p>
                )}
                {isEditing && agentIdStatus.availability && (
                  <p className={`text-sm ${
                    agentIdStatus.availability.available 
                      ? 'text-green-600' 
                      : 'text-destructive'
                  }`}>
                    {agentIdStatus.availability.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                {isEditing ? (
                  <Input
                    id="name"
                    value={editForm.name}
                    onChange={(e) => setEditForm(prev => ({ ...prev, name: e.target.value }))}
                  />
                ) : (
                  <p className="px-3 py-2 bg-muted rounded-md">{agent.name}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                {isEditing ? (
                  <Textarea
                    id="description"
                    value={editForm.description}
                    onChange={(e) => setEditForm(prev => ({ ...prev, description: e.target.value }))}
                    rows={3}
                  />
                ) : (
                  <p className="px-3 py-2 bg-muted rounded-md min-h-[80px]">{agent.description}</p>
                )}
              </div>
            </CardContent>
          </Card>

          {agent.latestVersion && (
            <Card>
              <CardHeader>
                <CardTitle>Current Configuration</CardTitle>
                <CardDescription>
                  Latest version settings (v{agent.latest})
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>LLM Provider</Label>
                    {isEditing ? (
                      <Select 
                        value={editForm.llmProvider} 
                        onValueChange={handleProviderChange}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {LLM_PROVIDERS.map(provider => (
                            <SelectItem key={provider.value} value={provider.value}>
                              {provider.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <p className="px-3 py-2 bg-muted rounded-md">{agent.latestVersion.llmProvider}</p>
                    )}
                  </div>
                  <div className="space-y-2">
                    <Label>Model</Label>
                    {isEditing ? (
                      <Select 
                        value={editForm.llmModel} 
                        onValueChange={(value: string) => setEditForm(prev => ({ ...prev, llmModel: value }))}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {COMMON_MODELS[editForm.llmProvider].map(model => (
                            <SelectItem key={model} value={model}>
                              {model}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <p className="px-3 py-2 bg-muted rounded-md">{agent.latestVersion.llmModel}</p>
                    )}
                  </div>
                </div>

                <div className="space-y-2">
                  <Label>System Prompt</Label>
                  {isEditing ? (
                    <Textarea
                      value={editForm.systemPrompt}
                      onChange={(e) => setEditForm(prev => ({ ...prev, systemPrompt: e.target.value }))}
                      rows={6}
                      className="font-mono"
                      placeholder="Enter the system prompt for this agent..."
                    />
                  ) : (
                    <Textarea
                      value={agent.latestVersion.systemPrompt}
                      readOnly
                      rows={6}
                      className="font-mono bg-muted"
                    />
                  )}
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <User className="h-5 w-5" />
                <span>Metadata</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>Author</Label>
                <p className="text-sm text-muted-foreground">{agent.author}</p>
              </div>
              
              <div className="space-y-2">
                <Label>Status</Label>
                <div>
                  <Badge variant={agent.status === 'ACTIVE' ? 'default' : 'secondary'} className="text-xs">
                    {agent.status === 'ACTIVE' ? 'Active' : 'Archived'}
                  </Badge>
                </div>
              </div>
              
              <div className="space-y-2">
                <Label>Latest Version</Label>
                <p className="text-sm text-muted-foreground">v{agent.latest}</p>
              </div>

              <Separator />

              <div className="space-y-2">
                <Label className="flex items-center space-x-2">
                  <Clock className="h-4 w-4" />
                  <span>Created</span>
                </Label>
                <p className="text-sm text-muted-foreground">
                  {new Date(agent.createdAt).toLocaleString()}
                </p>
              </div>

              <div className="space-y-2">
                <Label className="flex items-center space-x-2">
                  <Clock className="h-4 w-4" />
                  <span>Updated</span>
                </Label>
                <p className="text-sm text-muted-foreground">
                  {new Date(agent.updatedAt).toLocaleString()}
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <History className="h-5 w-5" />
                <span>Actions</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Button 
                variant="outline" 
                className="w-full justify-start"
                onClick={() => navigate(`/agents/${agent.id}/versions`)}
              >
                <History className="mr-2 h-4 w-4" />
                Version History
              </Button>
              
              <Button 
                variant="outline" 
                className="w-full justify-start"
                onClick={() => setShowCreateVersionDialog(true)}
              >
                <Edit3 className="mr-2 h-4 w-4" />
                Create New Version
              </Button>

              <Separator />

              {agent.status === 'ACTIVE' ? (
                <Button 
                  variant="destructive" 
                  className="w-full justify-start"
                  disabled={isUpdatingStatus}
                  onClick={() => setShowArchiveDialog(true)}
                >
                  {isUpdatingStatus ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Archive className="mr-2 h-4 w-4" />
                  )}
                  {isUpdatingStatus ? 'Archiving...' : 'Archive Agent'}
                </Button>
              ) : (
                <Button 
                  variant="outline" 
                  className="w-full justify-start"
                  disabled={isUpdatingStatus}
                  onClick={() => setShowUnarchiveDialog(true)}
                >
                  {isUpdatingStatus ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <ArchiveRestore className="mr-2 h-4 w-4" />
                  )}
                  {isUpdatingStatus ? 'Unarchiving...' : 'Unarchive Agent'}
                </Button>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {agent && (
        <CreateVersionDialog
          agent={agent}
          open={showCreateVersionDialog}
          onOpenChange={setShowCreateVersionDialog}
          onVersionCreated={() => {
            // Refresh agent data to get updated version info
            fetchAgent()
          }}
        />
      )}

      {/* Archive Confirmation Dialog */}
      {agent && (
        <ConfirmDialog
          open={showArchiveDialog}
          onOpenChange={setShowArchiveDialog}
          title="Archive Agent"
          description={`Are you sure you want to archive "${agent.name}"? Archived agents cannot be used in Slack conversations but can be restored later.`}
          confirmText="Archive"
          cancelText="Cancel"
          confirmVariant="destructive"
          onConfirm={handleArchive}
        />
      )}

      {/* Unarchive Confirmation Dialog */}
      {agent && (
        <ConfirmDialog
          open={showUnarchiveDialog}
          onOpenChange={setShowUnarchiveDialog}
          title="Unarchive Agent"
          description={`Are you sure you want to unarchive "${agent.name}"? This will make the agent available for use in Slack conversations again.`}
          confirmText="Unarchive"
          cancelText="Cancel"
          confirmVariant="default"
          onConfirm={handleUnarchive}
        />
      )}
    </div>
  )
}