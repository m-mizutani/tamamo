import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Loader2, AlertCircle, CheckCircle } from 'lucide-react'
import { 
  Agent, 
  CREATE_AGENT_VERSION, 
  graphqlRequest,
  CreateAgentVersionInput
} from '@/lib/graphql'

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

interface CreateVersionDialogProps {
  agent: Agent
  open: boolean
  onOpenChange: (open: boolean) => void
  onVersionCreated?: () => void
}

export function CreateVersionDialog({ 
  agent, 
  open, 
  onOpenChange, 
  onVersionCreated 
}: CreateVersionDialogProps) {
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const [formData, setFormData] = useState({
    version: '',
    systemPrompt: agent.latestVersion?.systemPrompt || '',
    llmProvider: (agent.latestVersion?.llmProvider || 'OPENAI') as 'OPENAI' | 'CLAUDE' | 'GEMINI',
    llmModel: agent.latestVersion?.llmModel || 'gpt-4o'
  })

  const [validationErrors, setValidationErrors] = useState({
    version: '',
    systemPrompt: '',
    llmModel: ''
  })

  const validateForm = () => {
    const errors = {
      version: '',
      systemPrompt: '',
      llmModel: ''
    }

    // Version validation (strict semantic versioning - Major.Minor.Patch only)
    if (!formData.version.trim()) {
      errors.version = 'Version is required'
    } else if (!/^\d+\.\d+\.\d+$/.test(formData.version.trim())) {
      errors.version = 'Version must follow semantic versioning (e.g., 1.0.0)'
    }

    // System prompt validation (optional)
    if (formData.systemPrompt.trim() && formData.systemPrompt.trim().length < 10) {
      errors.systemPrompt = 'System prompt must be at least 10 characters if provided'
    }

    // Model validation
    if (!formData.llmModel.trim()) {
      errors.llmModel = 'Model is required'
    }

    setValidationErrors(errors)
    return !Object.values(errors).some(error => error !== '')
  }

  const handleSubmit = async () => {
    if (!validateForm()) {
      return
    }

    try {
      setCreating(true)
      setError(null)
      setSuccess(false)

      const input: CreateAgentVersionInput = {
        agentUuid: agent.id,
        version: formData.version.trim(),
        systemPrompt: formData.systemPrompt.trim() || undefined,
        llmProvider: formData.llmProvider,
        llmModel: formData.llmModel.trim()
      }

      await graphqlRequest<{ createAgentVersion: any }>(CREATE_AGENT_VERSION, {
        input
      })

      setSuccess(true)
      
      // Call callback if provided
      if (onVersionCreated) {
        onVersionCreated()
      }

      // Auto-close dialog after successful creation
      setTimeout(() => {
        onOpenChange(false)
        setSuccess(false)
        // Reset form
        setFormData(prev => ({
          ...prev,
          version: '',
          systemPrompt: agent.latestVersion?.systemPrompt || ''
        }))
      }, 1500)

    } catch (err) {
      console.error('Failed to create agent version:', err)
      setError(err instanceof Error ? err.message : 'Failed to create version')
    } finally {
      setCreating(false)
    }
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!creating) {
      onOpenChange(newOpen)
      if (!newOpen) {
        setError(null)
        setSuccess(false)
        setValidationErrors({ version: '', systemPrompt: '', llmModel: '' })
      }
    }
  }

  const handleProviderChange = (provider: 'OPENAI' | 'CLAUDE' | 'GEMINI') => {
    setFormData(prev => ({
      ...prev,
      llmProvider: provider,
      llmModel: COMMON_MODELS[provider][0] // Set first model as default
    }))
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Version</DialogTitle>
          <DialogDescription>
            Create a new version for <strong>{agent.name}</strong> ({agent.agentId})
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {error && (
            <div className="flex items-center space-x-2 text-destructive bg-destructive/10 p-3 rounded-md border border-destructive/20">
              <AlertCircle className="h-4 w-4" />
              <span className="text-sm">{error}</span>
            </div>
          )}

          {success && (
            <div className="flex items-center space-x-2 text-green-600 bg-green-50 p-3 rounded-md border border-green-200">
              <CheckCircle className="h-4 w-4" />
              <span className="text-sm">Version created successfully!</span>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="version">Version *</Label>
            <Input
              id="version"
              value={formData.version}
              onChange={(e) => setFormData(prev => ({ ...prev, version: e.target.value }))}
              placeholder="e.g., 1.0.0"
              disabled={creating}
              className={validationErrors.version ? 'border-destructive' : ''}
            />
            {validationErrors.version && (
              <p className="text-sm text-destructive">{validationErrors.version}</p>
            )}
            <p className="text-xs text-muted-foreground">
              Current version: v{agent.latest}
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="provider">LLM Provider *</Label>
              <Select 
                value={formData.llmProvider} 
                onValueChange={handleProviderChange}
                disabled={creating}
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
            </div>

            <div className="space-y-2">
              <Label htmlFor="model">Model *</Label>
              <Select 
                value={formData.llmModel} 
                onValueChange={(value: string) => setFormData(prev => ({ ...prev, llmModel: value }))}
                disabled={creating}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {COMMON_MODELS[formData.llmProvider].map(model => (
                    <SelectItem key={model} value={model}>
                      {model}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {validationErrors.llmModel && (
                <p className="text-sm text-destructive">{validationErrors.llmModel}</p>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="systemPrompt">System Prompt</Label>
            <Textarea
              id="systemPrompt"
              value={formData.systemPrompt}
              onChange={(e) => setFormData(prev => ({ ...prev, systemPrompt: e.target.value }))}
              placeholder="Enter the system prompt for this version..."
              rows={8}
              disabled={creating}
              className={validationErrors.systemPrompt ? 'border-destructive' : ''}
            />
            {validationErrors.systemPrompt && (
              <p className="text-sm text-destructive">{validationErrors.systemPrompt}</p>
            )}
            <p className="text-xs text-muted-foreground">
              {formData.systemPrompt.length} characters
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button 
            variant="outline" 
            onClick={() => handleOpenChange(false)}
            disabled={creating}
          >
            Cancel
          </Button>
          <Button 
            onClick={handleSubmit}
            disabled={creating || success}
          >
            {creating ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating...
              </>
            ) : success ? (
              <>
                <CheckCircle className="mr-2 h-4 w-4" />
                Created!
              </>
            ) : (
              'Create Version'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}