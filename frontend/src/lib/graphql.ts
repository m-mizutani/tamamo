// GraphQL queries and mutations for Agent management

export const GET_AGENTS = `
  query GetAgents($offset: Int, $limit: Int) {
    agents(offset: $offset, limit: $limit) {
      agents {
        id
        agentId
        name
        description
        author
        status
        latest
        createdAt
        updatedAt
        latestVersion {
          systemPrompt
          llmProvider
          llmModel
        }
      }
      totalCount
    }
  }
`;

export const GET_AGENTS_BY_STATUS = `
  query GetAgentsByStatus($status: AgentStatus!, $offset: Int, $limit: Int) {
    agentsByStatus(status: $status, offset: $offset, limit: $limit) {
      agents {
        id
        agentId
        name
        description
        author
        status
        latest
        createdAt
        updatedAt
        latestVersion {
          systemPrompt
          llmProvider
          llmModel
        }
      }
      totalCount
    }
  }
`;

export const GET_ALL_AGENTS = `
  query GetAllAgents($offset: Int, $limit: Int) {
    allAgents(offset: $offset, limit: $limit) {
      agents {
        id
        agentId
        name
        description
        author
        status
        latest
        createdAt
        updatedAt
        latestVersion {
          systemPrompt
          llmProvider
          llmModel
        }
      }
      totalCount
    }
  }
`;

export const GET_AGENT = `
  query GetAgent($id: ID!) {
    agent(id: $id) {
      id
      agentId
      name
      description
      author
      status
      latest
      createdAt
      updatedAt
      latestVersion {
        systemPrompt
        llmProvider
        llmModel
      }
    }
  }
`;

export const CHECK_AGENT_ID_AVAILABILITY = `
  query CheckAgentIdAvailability($agentId: String!) {
    checkAgentIdAvailability(agentId: $agentId) {
      available
      message
    }
  }
`;

export const CREATE_AGENT = `
  mutation CreateAgent($input: CreateAgentInput!) {
    createAgent(input: $input) {
      id
      agentId
      name
      description
      author
      status
      latest
      createdAt
      updatedAt
      latestVersion {
        systemPrompt
        llmProvider
        llmModel
      }
    }
  }
`;

export const UPDATE_AGENT = `
  mutation UpdateAgent($id: ID!, $input: UpdateAgentInput!) {
    updateAgent(id: $id, input: $input) {
      id
      agentId
      name
      description
      author
      status
      latest
      createdAt
      updatedAt
      latestVersion {
        systemPrompt
        llmProvider
        llmModel
      }
    }
  }
`;

export const DELETE_AGENT = `
  mutation DeleteAgent($id: ID!) {
    deleteAgent(id: $id)
  }
`;

export const ARCHIVE_AGENT = `
  mutation ArchiveAgent($id: ID!) {
    archiveAgent(id: $id) {
      id
      agentId
      name
      description
      author
      status
      latest
      createdAt
      updatedAt
      latestVersion {
        systemPrompt
        llmProvider
        llmModel
      }
    }
  }
`;

export const UNARCHIVE_AGENT = `
  mutation UnarchiveAgent($id: ID!) {
    unarchiveAgent(id: $id) {
      id
      agentId
      name
      description
      author
      status
      latest
      createdAt
      updatedAt
      latestVersion {
        systemPrompt
        llmProvider
        llmModel
      }
    }
  }
`;

export const CREATE_AGENT_VERSION = `
  mutation CreateAgentVersion($input: CreateAgentVersionInput!) {
    createAgentVersion(input: $input) {
      agentUuid
      version
      systemPrompt
      llmProvider
      llmModel
      createdAt
      updatedAt
    }
  }
`;

export const GET_AGENT_VERSIONS = `
  query GetAgentVersions($agentUuid: ID!) {
    agentVersions(agentUuid: $agentUuid) {
      agentUuid
      version
      systemPrompt
      llmProvider
      llmModel
      createdAt
      updatedAt
    }
  }
`;

// Type definitions
export type AgentStatus = 'ACTIVE' | 'ARCHIVED';

export interface Agent {
  id: string;
  agentId: string;
  name: string;
  description: string;
  author: string;
  status: AgentStatus;
  latest: string;
  createdAt: string;
  updatedAt: string;
  latestVersion?: AgentVersion;
}

export interface AgentVersion {
  agentUuid: string;
  version: string;
  systemPrompt: string;
  llmProvider: 'OPENAI' | 'CLAUDE' | 'GEMINI';
  llmModel: string;
  createdAt: string;
  updatedAt: string;
}

export interface AgentListResponse {
  agents: Agent[];
  totalCount: number;
}

export interface AgentIdAvailability {
  available: boolean;
  message: string;
}

export interface CreateAgentInput {
  agentId: string;
  name: string;
  description?: string;
  systemPrompt?: string;
  llmProvider: 'OPENAI' | 'CLAUDE' | 'GEMINI';
  llmModel: string;
  version?: string;
}

export interface UpdateAgentInput {
  agentId?: string;
  name?: string;
  description?: string;
  systemPrompt?: string;
  llmProvider?: 'OPENAI' | 'CLAUDE' | 'GEMINI';
  llmModel?: string;
}

export interface CreateAgentVersionInput {
  agentUuid: string;
  version: string;
  systemPrompt?: string;
  llmProvider: 'OPENAI' | 'CLAUDE' | 'GEMINI';
  llmModel: string;
}

// GraphQL client utility
export async function graphqlRequest<T>(
  query: string,
  variables?: Record<string, any>,
  signal?: AbortSignal
): Promise<T> {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
    body: JSON.stringify({
      query,
      variables,
    }),
    signal,
  });

  // Handle authentication errors
  if (response.status === 401) {
    // Clear any cached auth state and redirect to login
    window.location.href = '/api/auth/login';
    throw new Error('Authentication required');
  }

  if (!response.ok) {
    throw new Error(`GraphQL request failed: ${response.statusText}`);
  }

  const result = await response.json();

  if (result.errors) {
    // Check if any error is an authentication error
    const hasAuthError = result.errors.some((e: any) => 
      e.message?.toLowerCase().includes('unauthorized') ||
      e.message?.toLowerCase().includes('authentication')
    );
    
    if (hasAuthError) {
      window.location.href = '/api/auth/login';
      throw new Error('Authentication required');
    }
    
    throw new Error(`GraphQL errors: ${result.errors.map((e: any) => e.message).join(', ')}`);
  }

  return result.data;
}