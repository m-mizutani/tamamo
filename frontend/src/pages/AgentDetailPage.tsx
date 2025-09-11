import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Separator } from '@/components/ui/separator'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ImageUpload } from '@/components/ImageUpload'
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
  Settings,
  Image as ImageIcon
} from 'lucide-react'
import { 
  Agent,
  GET_AGENT, 
  UPDATE_AGENT,
  CHECK_AGENT_ID_AVAILABILITY,
  ARCHIVE_AGENT,
  UNARCHIVE_AGENT,
  GET_LLM_CONFIG,
  graphqlRequest,
  AgentIdAvailability,
  UpdateAgentInput,
  LLMConfig
} from '@/lib/graphql'
import { useImageUpload } from '@/hooks/useImageUpload'
import { Badge } from '@/components/ui/badge'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { CreateVersionDialog } from '@/components/agents/CreateVersionDialog'
import { UserDisplay } from '@/components/UserDisplay'
import { AgentSearchConfigSection } from '@/components/agent/search-config/AgentSearchConfigSection'

// Helper functions to get display names
function getProviderDisplayName(providerId: string | undefined, llmConfig: LLMConfig | null): string {
  if (!providerId || !llmConfig) return providerId || 'Unknown'
  const provider = llmConfig.providers.find(p => p.id === providerId.toLowerCase())
  return provider?.displayName || providerId
}

function getModelDisplayName(providerId: string | undefined, modelId: string | undefined, llmConfig: LLMConfig | null): string {
  if (!providerId || !modelId || !llmConfig) return modelId || 'Unknown'
  const provider = llmConfig.providers.find(p => p.id === providerId.toLowerCase())
  const model = provider?.models.find(m => m.id === modelId)
  return model?.displayName || modelId
}

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
  const [llmConfig, setLlmConfig] = useState<LLMConfig | null>(null)
  
  // Image upload hook
  const imageUpload = useImageUpload({
    initialImageUrl: agent?.imageUrl,
    onSuccess: () => {
      // Refresh agent data after successful image upload
      fetchAgent()
    },
    onError: (error) => setError(`Image upload failed: ${error}`)
  })
  
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
      setError(err instanceof Error ? err.message : 'Failed to load agent')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const controller = new AbortController()
    fetchAgent(controller.signal)
    
    // Fetch LLM configuration
    graphqlRequest<{ llmConfig: LLMConfig }>(GET_LLM_CONFIG)
      .then(response => setLlmConfig(response.llmConfig))
      .catch(() => {
        // Silently handle LLM config fetch errors
      })
    
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
    imageUpload.reset()
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
      
      // Upload image if one was selected
      if (imageUpload.selectedFile && agent) {
        try {
          await imageUpload.uploadImage(agent.id)
        } catch (imageError) {
          // Don't prevent saving the agent if image upload fails
        }
      }
      
      setIsEditing(false)
    } catch (err) {
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
    // Find the provider config and set the first available model
    const providerConfig = llmConfig?.providers.find(p => p.id === provider.toLowerCase())
    const firstModel = providerConfig?.models[0]?.id || ''
    
    setEditForm(prev => ({
      ...prev,
      llmProvider: provider,
      llmModel: firstModel
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
      setError(err instanceof Error ? err.message : 'Failed to unarchive agent')
    } finally {
      setIsUpdatingStatus(false)
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center">
          <Button variant="ghost" size="sm" onClick={() => navigate('/agents')} className="text-muted-foreground hover:text-foreground">
            <ArrowLeft className="mr-1 h-3 w-3" />
            <span className="text-xs">Back to Agents</span>
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
        <div className="flex items-center">
          <Button variant="ghost" size="sm" onClick={() => navigate('/agents')} className="text-muted-foreground hover:text-foreground">
            <ArrowLeft className="mr-1 h-3 w-3" />
            <span className="text-xs">Back to Agents</span>
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
        <div className="flex items-center">
          <Button variant="ghost" size="sm" onClick={() => navigate('/agents')} className="text-muted-foreground hover:text-foreground">
            <ArrowLeft className="mr-1 h-3 w-3" />
            <span className="text-xs">Back to Agents</span>
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
      {/* Small Back Button */}
      <div className="flex items-center">
        <Button variant="ghost" size="sm" onClick={() => navigate('/agents')} className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="mr-1 h-3 w-3" />
          <span className="text-xs">Back to Agents</span>
        </Button>
      </div>
      
      {/* Main Agent Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center space-x-3">
            <h1 className="text-3xl font-bold tracking-tight">{agent.name}</h1>
            <Badge variant={agent.status === 'ACTIVE' ? 'default' : 'secondary'}>
              {agent.status === 'ACTIVE' ? 'Active' : 'Archived'}
            </Badge>
          </div>
          <p className="text-muted-foreground mt-1">
            {agent.agentId} • v{agent.latest}
          </p>
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
                  Latest version settings{agent.latest && ` (v${agent.latest})`}
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
                          {llmConfig?.providers.map(provider => (
                            <SelectItem key={provider.id} value={provider.id.toUpperCase()}>
                              {provider.displayName}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <p className="px-3 py-2 bg-muted rounded-md">
                        {getProviderDisplayName(agent.latestVersion.llmProvider, llmConfig)}
                      </p>
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
                          {llmConfig?.providers
                            .find(p => p.id === editForm.llmProvider.toLowerCase())
                            ?.models.map(model => (
                              <SelectItem key={model.id} value={model.id}>
                                {model.displayName}
                                {model.description && (
                                  <span className="text-xs text-muted-foreground ml-2">
                                    {model.description}
                                  </span>
                                )}
                              </SelectItem>
                            ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <p className="px-3 py-2 bg-muted rounded-md">
                        {getModelDisplayName(agent.latestVersion.llmProvider, agent.latestVersion.llmModel, llmConfig)}
                      </p>
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

          <AgentSearchConfigSection agentId={agent.id} canEdit={true} />
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <ImageIcon className="h-5 w-5" />
                <span>Agent Image</span>
              </CardTitle>
              <CardDescription>
                Image representing this agent
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isEditing ? (
                <ImageUpload
                  onImageSelect={imageUpload.handleFileSelect}
                  previewUrl={imageUpload.preview}
                  isUploading={imageUpload.isUploading}
                  error={imageUpload.error}
                />
              ) : (
                <div className="space-y-4">
                  {agent?.imageUrl ? (
                    <div className="flex flex-col items-center space-y-2">
                      <div className="w-32 h-32 rounded-lg overflow-hidden border-2 border-gray-200">
                        <img
                          src={agent.imageUrl}
                          alt={`${agent.name} image`}
                          className="w-full h-full object-cover"
                        />
                      </div>
                      <p className="text-sm text-muted-foreground">Current agent image</p>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center space-y-2 py-8">
                      <ImageIcon className="h-12 w-12 text-muted-foreground" />
                      <p className="text-sm text-muted-foreground">No image uploaded</p>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>

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
                <UserDisplay user={agent.author} size={24} />
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
                {agent.latest && <p className="text-sm text-muted-foreground">v{agent.latest}</p>}
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