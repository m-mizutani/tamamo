import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { 
  History, 
  Clock, 
  Settings,
  ChevronDown,
  ChevronRight,
  Loader2,
  AlertCircle
} from 'lucide-react'
import { 
  AgentVersion, 
  GET_AGENT_VERSIONS, 
  graphqlRequest 
} from '@/lib/graphql'

interface VersionHistoryProps {
  agentUuid: string
  currentVersion?: string
}

export function VersionHistory({ agentUuid, currentVersion }: VersionHistoryProps) {
  const [versions, setVersions] = useState<AgentVersion[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedVersions, setExpandedVersions] = useState<Set<string>>(new Set())

  const fetchVersions = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await graphqlRequest<{ agentVersions: AgentVersion[] }>(
        GET_AGENT_VERSIONS, 
        { agentUuid }
      )
      setVersions(response.agentVersions)
    } catch (err) {
      console.error('Failed to fetch agent versions:', err)
      setError(err instanceof Error ? err.message : 'Failed to load versions')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchVersions()
  }, [agentUuid])

  const toggleVersionExpansion = (version: string) => {
    setExpandedVersions(prev => {
      const newSet = new Set(prev)
      if (newSet.has(version)) {
        newSet.delete(version)
      } else {
        newSet.add(version)
      }
      return newSet
    })
  }

  const getLLMProviderColor = (provider: string | undefined) => {
    if (!provider) return 'bg-gray-100 text-gray-800 border-gray-200'
    switch (provider) {
      case 'OPENAI': return 'bg-green-100 text-green-800 border-green-200'
      case 'CLAUDE': return 'bg-orange-100 text-orange-800 border-orange-200'
      case 'GEMINI': return 'bg-blue-100 text-blue-800 border-blue-200'
      default: return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <History className="h-5 w-5" />
            <span>Version History</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            <span className="ml-2 text-muted-foreground">Loading versions...</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <History className="h-5 w-5" />
            <span>Version History</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8 text-destructive">
            <AlertCircle className="h-5 w-5 mr-2" />
            <span>{error}</span>
          </div>
          <div className="text-center mt-4">
            <Button onClick={fetchVersions} variant="outline" size="sm">
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <History className="h-5 w-5" />
          <span>Version History</span>
        </CardTitle>
        <CardDescription>
          {versions.length} version{versions.length !== 1 ? 's' : ''} available
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {versions.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <History className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p>No versions found</p>
          </div>
        ) : (
          versions.map((version, index) => {
            const isExpanded = expandedVersions.has(version.version)
            const isCurrent = version.version === currentVersion
            
            return (
              <div key={version.version} className="space-y-2">
                {index > 0 && <Separator />}
                
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => toggleVersionExpansion(version.version)}
                        className="h-8 w-8 p-0"
                      >
                        {isExpanded ? (
                          <ChevronDown className="h-4 w-4" />
                        ) : (
                          <ChevronRight className="h-4 w-4" />
                        )}
                      </Button>
                      
                      <div className="flex items-center space-x-2">
                        <span className="font-medium">v{version.version}</span>
                        {isCurrent && (
                          <Badge variant="default" className="text-xs">
                            Current
                          </Badge>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                      <Clock className="h-4 w-4" />
                      <span>{new Date(version.createdAt).toLocaleDateString()}</span>
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="ml-11 space-y-4 p-4 bg-muted/50 rounded-lg">
                      <div className="grid grid-cols-2 gap-4">
                        {version.llmProvider && (
                          <div className="space-y-1">
                            <p className="text-sm font-medium">LLM Provider</p>
                            <Badge 
                              variant="outline" 
                              className={getLLMProviderColor(version.llmProvider)}
                            >
                              {version.llmProvider}
                            </Badge>
                          </div>
                        )}
                        {version.llmModel && (
                          <div className="space-y-1">
                            <p className="text-sm font-medium">Model</p>
                            <p className="text-sm text-muted-foreground">{version.llmModel}</p>
                          </div>
                        )}
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-center space-x-2">
                          <Settings className="h-4 w-4" />
                          <p className="text-sm font-medium">System Prompt</p>
                        </div>
                        <div className="bg-background border rounded-md p-3">
                          <pre className="text-xs text-muted-foreground whitespace-pre-wrap font-mono">
                            {version.systemPrompt}
                          </pre>
                        </div>
                      </div>

                      <div className="flex items-center space-x-4 text-xs text-muted-foreground">
                        <span>Created: {new Date(version.createdAt).toLocaleString()}</span>
                        <span>•</span>
                        <span>Updated: {new Date(version.updatedAt).toLocaleString()}</span>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )
          })
        )}
      </CardContent>
    </Card>
  )
}