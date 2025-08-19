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
export interface Agent {
  id: string;
  agentId: string;
  name: string;
  description: string;
  author: string;
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
  description: string;
  systemPrompt: string;
  llmProvider: 'OPENAI' | 'CLAUDE' | 'GEMINI';
  llmModel: string;
  version?: string;
}

export interface UpdateAgentInput {
  agentId?: string;
  name?: string;
  description?: string;
}

export interface CreateAgentVersionInput {
  agentUuid: string;
  version: string;
  systemPrompt: string;
  llmProvider: 'OPENAI' | 'CLAUDE' | 'GEMINI';
  llmModel: string;
}

// GraphQL client utility
export async function graphqlRequest<T>(
  query: string,
  variables?: Record<string, any>
): Promise<T> {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query,
      variables,
    }),
  });

  if (!response.ok) {
    throw new Error(`GraphQL request failed: ${response.statusText}`);
  }

  const result = await response.json();

  if (result.errors) {
    throw new Error(`GraphQL errors: ${result.errors.map((e: any) => e.message).join(', ')}`);
  }

  return result.data;
}