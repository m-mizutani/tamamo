import { useState, useCallback } from 'react'
import { graphqlRequest } from '@/lib/graphql'
import { toast } from 'sonner'
import {
  AgentNotionSearchConfig,
  CreateNotionSearchConfigInput,
  UpdateNotionSearchConfigInput,
  GET_AGENT_NOTION_SEARCH_CONFIGS,
  CREATE_NOTION_SEARCH_CONFIG,
  UPDATE_NOTION_SEARCH_CONFIG,
  DELETE_NOTION_SEARCH_CONFIG
} from '@/lib/graphql'

export function useNotionSearchConfigs(agentId: string) {
  const [configs, setConfigs] = useState<AgentNotionSearchConfig[]>([])
  const [loading, setLoading] = useState(false)

  const loadConfigs = useCallback(async () => {
    try {
      setLoading(true)
      const data = await graphqlRequest<{agentNotionSearchConfigs: AgentNotionSearchConfig[]}>(
        GET_AGENT_NOTION_SEARCH_CONFIGS,
        { agentId }
      )
      setConfigs(data.agentNotionSearchConfigs || [])
    } catch (error) {
      console.error('Failed to load Notion search configs:', error)
      toast.error('Failed to load Notion search configurations')
    } finally {
      setLoading(false)
    }
  }, [agentId])

  const createConfig = useCallback(async (input: CreateNotionSearchConfigInput) => {
    try {
      const data = await graphqlRequest<{createNotionSearchConfig: AgentNotionSearchConfig}>(
        CREATE_NOTION_SEARCH_CONFIG,
        { input }
      )
      setConfigs(prev => [...prev, data.createNotionSearchConfig])
      toast.success('Notion search configuration added successfully')
      return data.createNotionSearchConfig
    } catch (error) {
      console.error('Failed to create Notion search config:', error)
      toast.error('Failed to add Notion search configuration')
      throw error
    }
  }, [])

  const updateConfig = useCallback(async (id: string, input: UpdateNotionSearchConfigInput) => {
    try {
      const data = await graphqlRequest<{updateNotionSearchConfig: AgentNotionSearchConfig}>(
        UPDATE_NOTION_SEARCH_CONFIG,
        { id, input }
      )
      setConfigs(prev => prev.map(config => 
        config.id === id ? data.updateNotionSearchConfig : config
      ))
      toast.success('Notion search configuration updated successfully')
      return data.updateNotionSearchConfig
    } catch (error) {
      console.error('Failed to update Notion search config:', error)
      toast.error('Failed to update Notion search configuration')
      throw error
    }
  }, [])

  const deleteConfig = useCallback(async (id: string) => {
    try {
      await graphqlRequest<{deleteNotionSearchConfig: boolean}>(
        DELETE_NOTION_SEARCH_CONFIG,
        { id }
      )
      setConfigs(prev => prev.filter(config => config.id !== id))
      toast.success('Notion search configuration deleted successfully')
    } catch (error) {
      console.error('Failed to delete Notion search config:', error)
      toast.error('Failed to delete Notion search configuration')
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