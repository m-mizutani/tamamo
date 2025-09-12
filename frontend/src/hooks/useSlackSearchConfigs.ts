import { useState, useCallback } from 'react'
import { graphqlRequest } from '@/lib/graphql'
import { toast } from 'sonner'
import {
  AgentSlackSearchConfig,
  CreateSlackSearchConfigInput,
  UpdateSlackSearchConfigInput,
  GET_AGENT_SLACK_SEARCH_CONFIGS,
  CREATE_SLACK_SEARCH_CONFIG,
  UPDATE_SLACK_SEARCH_CONFIG,
  DELETE_SLACK_SEARCH_CONFIG
} from '@/lib/graphql'

export function useSlackSearchConfigs(agentId: string) {
  const [configs, setConfigs] = useState<AgentSlackSearchConfig[]>([])
  const [loading, setLoading] = useState(false)

  const loadConfigs = useCallback(async () => {
    try {
      setLoading(true)
      const data = await graphqlRequest<{agentSlackSearchConfigs: AgentSlackSearchConfig[]}>(
        GET_AGENT_SLACK_SEARCH_CONFIGS,
        { agentId }
      )
      setConfigs(data.agentSlackSearchConfigs || [])
    } catch (error) {
      console.error('Failed to load Slack search configs:', error)
      toast.error('Failed to load Slack search configurations')
    } finally {
      setLoading(false)
    }
  }, [agentId])

  const createConfig = useCallback(async (input: CreateSlackSearchConfigInput) => {
    try {
      const data = await graphqlRequest<{createSlackSearchConfig: AgentSlackSearchConfig}>(
        CREATE_SLACK_SEARCH_CONFIG,
        { input }
      )
      setConfigs(prev => [...prev, data.createSlackSearchConfig])
      toast.success('Slack search configuration added successfully')
      return data.createSlackSearchConfig
    } catch (error) {
      console.error('Failed to create Slack search config:', error)
      toast.error('Failed to add Slack search configuration')
      throw error
    }
  }, [])

  const updateConfig = useCallback(async (id: string, input: UpdateSlackSearchConfigInput) => {
    try {
      const data = await graphqlRequest<{updateSlackSearchConfig: AgentSlackSearchConfig}>(
        UPDATE_SLACK_SEARCH_CONFIG,
        { id, input }
      )
      setConfigs(prev => prev.map(config => 
        config.id === id ? data.updateSlackSearchConfig : config
      ))
      toast.success('Slack search configuration updated successfully')
      return data.updateSlackSearchConfig
    } catch (error) {
      console.error('Failed to update Slack search config:', error)
      toast.error('Failed to update Slack search configuration')
      throw error
    }
  }, [])

  const deleteConfig = useCallback(async (id: string) => {
    try {
      await graphqlRequest<{deleteSlackSearchConfig: boolean}>(
        DELETE_SLACK_SEARCH_CONFIG,
        { id }
      )
      setConfigs(prev => prev.filter(config => config.id !== id))
      toast.success('Slack search configuration deleted successfully')
    } catch (error) {
      console.error('Failed to delete Slack search config:', error)
      toast.error('Failed to delete Slack search configuration')
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