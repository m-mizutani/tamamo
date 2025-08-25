import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ImageUpload } from '@/components/ImageUpload'
import { ArrowLeft, Save, Loader2, AlertCircle, CheckCircle } from 'lucide-react'
import { 
  CreateAgentInput, 
  CREATE_AGENT, 
  CHECK_AGENT_ID_AVAILABILITY, 
  GET_LLM_CONFIG,
  graphqlRequest,
  AgentIdAvailability,
  LLMConfig 
} from '@/lib/graphql'
import { useImageUpload } from '@/hooks/useImageUpload'

export function CreateAgentPage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [llmConfig, setLlmConfig] = useState<LLMConfig | null>(null)
  const [agentIdStatus, setAgentIdStatus] = useState<{
    checking: boolean
    availability: AgentIdAvailability | null
  }>({ checking: false, availability: null })

  // Image upload hook
  const imageUpload = useImageUpload({
    onError: (error) => setError(`Image upload failed: ${error}`)
  })

  const [formData, setFormData] = useState<CreateAgentInput>({
    agentId: '',
    name: '',
    description: '',
    systemPrompt: '',
    llmProvider: undefined,  // No hardcoded default
    llmModel: undefined,
    version: '1.0.0'
  })

  // Fetch LLM configuration on mount
  useEffect(() => {
    graphqlRequest<{ llmConfig: LLMConfig }>(GET_LLM_CONFIG)
      .then(response => {
        setLlmConfig(response.llmConfig)
        // Set default provider and model from config
        if (response.llmConfig.defaultProvider && response.llmConfig.defaultModel) {
          setFormData(prev => ({
            ...prev,
            llmProvider: response.llmConfig.defaultProvider.toUpperCase() as 'OPENAI' | 'CLAUDE' | 'GEMINI',
            llmModel: response.llmConfig.defaultModel
          }))
        }
      })
      .catch(err => console.error('Failed to fetch LLM config:', err))
  }, [])

  const checkAgentIdAvailability = async (agentId: string) => {
    if (!agentId || agentId.length < 2) {
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

  const handleAgentIdChange = (value: string) => {
    setFormData(prev => ({ ...prev, agentId: value }))
    
    // Debounce the availability check
    const timeoutId = setTimeout(() => {
      checkAgentIdAvailability(value)
    }, 500)
    
    return () => clearTimeout(timeoutId)
  }

  const handleAgentIdBlur = () => {
    if (formData.agentId) {
      checkAgentIdAvailability(formData.agentId)
    }
  }

  const handleProviderChange = (provider: 'OPENAI' | 'CLAUDE' | 'GEMINI') => {
    // Find the provider config and set the first available model
    const providerConfig = llmConfig?.providers.find(p => p.id === provider.toLowerCase())
    const firstModel = providerConfig?.models[0]?.id || ''
    
    setFormData(prev => ({
      ...prev,
      llmProvider: provider,
      llmModel: firstModel
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!agentIdStatus.availability?.available) {
      setError('Please choose an available Agent ID')
      return
    }

    try {
      setLoading(true)
      setError(null)
      
      // Prepare input, omitting empty optional fields
      const input: CreateAgentInput = {
        agentId: formData.agentId,
        name: formData.name,
        llmProvider: formData.llmProvider,
        llmModel: formData.llmModel,
        version: formData.version
      }
      
      // Add optional fields only if they have values
      if (formData.description?.trim()) {
        input.description = formData.description.trim()
      }
      if (formData.systemPrompt?.trim()) {
        input.systemPrompt = formData.systemPrompt.trim()
      }

      const response = await graphqlRequest<{ createAgent: any }>(CREATE_AGENT, {
        input
      })
      
      // Upload image if one was selected
      if (imageUpload.selectedFile) {
        try {
          await imageUpload.uploadImage(response.createAgent.id)
        } catch (imageError) {
          // Log image upload error but don't prevent navigation
          console.warn('Image upload failed:', imageError)
        }
      }
      
      // Navigate to the created agent's detail page
      navigate(`/agents/${response.createAgent.id}`)
    } catch (err) {
      console.error('Failed to create agent:', err)
      setError(err instanceof Error ? err.message : 'Failed to create agent')
    } finally {
      setLoading(false)
    }
  }

  const isFormValid = 
    formData.agentId &&
    formData.name &&
    formData.llmProvider &&
    formData.llmModel &&
    agentIdStatus.availability?.available

  return (
    <div className="space-y-6">
      <div className="flex items-center space-x-4">
        <Button 
          variant="ghost" 
          size="sm"
          onClick={() => navigate('/agents')}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Agents
        </Button>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create New Agent</h1>
          <p className="text-muted-foreground">
            Configure your AI agent with custom prompts and LLM settings
          </p>
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

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Basic Information</CardTitle>
            <CardDescription>
              Define the agent's identity and purpose
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="agentId">Agent ID *</Label>
              <div className="relative">
                <Input
                  id="agentId"
                  placeholder="my-agent"
                  value={formData.agentId}
                  onChange={(e) => handleAgentIdChange(e.target.value)}
                  onBlur={handleAgentIdBlur}
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
              {agentIdStatus.availability && (
                <p className={`text-sm ${
                  agentIdStatus.availability.available 
                    ? 'text-green-600' 
                    : 'text-destructive'
                }`}>
                  {agentIdStatus.availability.message}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Unique identifier using alphanumeric characters, hyphens, dots, and underscores
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="name">Name *</Label>
              <Input
                id="name"
                placeholder="Customer Support Agent"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe what this agent does and its purpose..."
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                rows={3}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Agent Image</CardTitle>
            <CardDescription>
              Upload an image to represent this agent (optional)
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ImageUpload
              onImageSelect={imageUpload.handleFileSelect}
              previewUrl={imageUpload.preview}
              isUploading={imageUpload.isUploading}
              error={imageUpload.error}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>LLM Configuration</CardTitle>
            <CardDescription>
              Choose the language model and provider for this agent
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="provider">LLM Provider *</Label>
                <Select value={formData.llmProvider} onValueChange={handleProviderChange}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a provider" />
                  </SelectTrigger>
                  <SelectContent>
                    {llmConfig?.providers.map(provider => (
                      <SelectItem key={provider.id} value={provider.id.toUpperCase()}>
                        {provider.displayName}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="model">Model *</Label>
                <Select 
                  value={formData.llmModel} 
                  onValueChange={(value: string) => setFormData(prev => ({ ...prev, llmModel: value }))}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a model" />
                  </SelectTrigger>
                  <SelectContent>
                    {llmConfig?.providers
                      .find(p => p.id === formData.llmProvider?.toLowerCase())
                      ?.models.map(model => (
                        <SelectItem key={model.id} value={model.id}>
                          {model.displayName}
                          {model.description && (
                            <span className="text-xs text-muted-foreground ml-2">
                              ({model.description})
                            </span>
                          )}
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Prompt</CardTitle>
            <CardDescription>
              Define the agent's behavior, personality, and instructions
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Textarea
              placeholder="You are a helpful AI assistant that..."
              value={formData.systemPrompt}
              onChange={(e) => setFormData(prev => ({ ...prev, systemPrompt: e.target.value }))}
              rows={8}
              className="font-mono"
            />
          </CardContent>
        </Card>

        <div className="flex justify-end space-x-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate('/agents')}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={!isFormValid || loading}
          >
            {loading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Create Agent
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  )
}