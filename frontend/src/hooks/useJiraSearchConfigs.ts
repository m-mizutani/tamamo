import { useState, useCallback } from 'react'
import { graphqlRequest } from '@/lib/graphql'
import { toast } from 'sonner'
import {
  AgentJiraSearchConfig,
  CreateJiraSearchConfigInput,
  UpdateJiraSearchConfigInput,
  GET_AGENT_JIRA_SEARCH_CONFIGS,
  CREATE_JIRA_SEARCH_CONFIG,
  UPDATE_JIRA_SEARCH_CONFIG,
  DELETE_JIRA_SEARCH_CONFIG
} from '@/lib/graphql'

export function useJiraSearchConfigs(agentId: string) {
  const [configs, setConfigs] = useState<AgentJiraSearchConfig[]>([])
  const [loading, setLoading] = useState(false)

  const loadConfigs = useCallback(async () => {
    try {
      setLoading(true)
      const data = await graphqlRequest<{agentJiraSearchConfigs: AgentJiraSearchConfig[]}>(
        GET_AGENT_JIRA_SEARCH_CONFIGS,
        { agentId }
      )
      setConfigs(data.agentJiraSearchConfigs || [])
    } catch (error) {
      console.error('Failed to load Jira search configs:', error)
      toast.error('Failed to load Jira search configurations')
    } finally {
      setLoading(false)
    }
  }, [agentId])

  const createConfig = useCallback(async (input: CreateJiraSearchConfigInput) => {
    try {
      const data = await graphqlRequest<{createJiraSearchConfig: AgentJiraSearchConfig}>(
        CREATE_JIRA_SEARCH_CONFIG,
        { input }
      )
      setConfigs(prev => [...prev, data.createJiraSearchConfig])
      toast.success('Jira search configuration added successfully')
      return data.createJiraSearchConfig
    } catch (error) {
      console.error('Failed to create Jira search config:', error)
      toast.error('Failed to add Jira search configuration')
      throw error
    }
  }, [])

  const updateConfig = useCallback(async (id: string, input: UpdateJiraSearchConfigInput) => {
    try {
      const data = await graphqlRequest<{updateJiraSearchConfig: AgentJiraSearchConfig}>(
        UPDATE_JIRA_SEARCH_CONFIG,
        { id, input }
      )
      setConfigs(prev => prev.map(config => 
        config.id === id ? data.updateJiraSearchConfig : config
      ))
      toast.success('Jira search configuration updated successfully')
      return data.updateJiraSearchConfig
    } catch (error) {
      console.error('Failed to update Jira search config:', error)
      toast.error('Failed to update Jira search configuration')
      throw error
    }
  }, [])

  const deleteConfig = useCallback(async (id: string) => {
    try {
      await graphqlRequest<{deleteJiraSearchConfig: boolean}>(
        DELETE_JIRA_SEARCH_CONFIG,
        { id }
      )
      setConfigs(prev => prev.filter(config => config.id !== id))
      toast.success('Jira search configuration deleted successfully')
    } catch (error) {
      console.error('Failed to delete Jira search config:', error)
      toast.error('Failed to delete Jira search configuration')
      throw error
    }
  }, [])

  return {
    configs,
    loading,
    loadConfigs,
    createConfig,
    updateConfig,
    deleteConfig
  }
}