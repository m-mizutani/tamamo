# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Tamamo is a Slack bot application built with Go, using Domain-Driven Design (DDD) and clean architecture principles. It provides an AI-powered assistant through Slack integration.

## Restrictions and Rules

### Directory

- When you are mentioned about `tmp` directory, you SHOULD NOT see `/tmp`. You need to check `./tmp` directory from root of the repository.

### Exposure policy

In principle, do not trust developers who use this library from outside

- Do not export unnecessary methods, structs, and variables
- Assume that exposed items will be changed. Never expose fields that would be problematic if changed
- Use `export_test.go` for items that need to be exposed for testing purposes
- **Exception**: Domain models (`pkg/domain/model/*`) can have exported fields as they represent data structures

### Firestore Struct Tags

- **NEVER use firestore struct tags on domain models**
- Domain models should be pure Go structs without any persistence-specific annotations
- This keeps the domain layer independent of the infrastructure layer

### Check

When making changes, before finishing the task, always:
- Run `go vet ./...`, `go fmt ./...` to format the code
- Run `golangci-lint run ./...` to check lint error
- Run `gosec -exclude-generated -quiet ./...` to check security issue
- Run `go test ./...` to check side effect
- **For GraphQL changes: Run `task graphql` and verify no compilation errors**
- **For GraphQL changes: Check frontend GraphQL queries are updated accordingly**

### Language

All comment and character literal in source code must be in English

### Testing

- Test files should have `package {name}_test`. Do not use same package name
- **🚨 CRITICAL RULE: Test MUST be included in same name test file. (e.g. test for `abc.go` must be in `abc_test.go`) 🚨**

#### Repository Testing Strategy
- **🚨 CRITICAL: Repository tests MUST be placed in `pkg/repository/database/` directory with common test suites**
- Create shared test functions that verify identical behavior across all repository implementations (Firestore, Memory, etc.)
- Each repository implementation must pass the exact same test suite to ensure behavioral consistency
- Use a common test interface pattern to test all implementations uniformly
- This ensures that switching between repository implementations (e.g., Memory for testing, Firestore for production) maintains identical behavior

### Slack Responses
- **All Slack responses MUST be sent as thread replies (with thread_ts)**
- Never send messages to the channel directly unless explicitly required

### GraphQL Development Workflow

**🚨 CRITICAL: GraphQL changes require a specific workflow to avoid runtime errors 🚨**

When modifying GraphQL schemas or domain models that affect GraphQL:

#### 1. Schema Changes
- Always update `graphql/schema.graphql` first
- Ensure all new fields have appropriate types and nullability
- Consider backward compatibility for existing clients

#### 2. Code Generation (MANDATORY)
- **ALWAYS run `task graphql` (or `task g`) after schema changes**
- This regenerates `pkg/controller/graphql/generated.go` and related files
- Generated code must be committed along with schema changes

#### 3. Domain Model Updates
- If adding fields to domain models, update all related structures:
  - Domain model struct (e.g., `pkg/domain/model/user/user.go`)
  - Repository layer (Firestore and Memory implementations)
  - Constructor functions and update methods
  - GraphQL interface types if used

#### 4. Testing Updates
- Update all test mocks to include new fields
- Update test data creation (e.g., `NewUser` calls)
- Update usecase constructor calls if signatures changed
- Run tests early and often during development

#### 5. Frontend Synchronization
- Update TypeScript interfaces in `frontend/src/lib/graphql.ts`
- Update GraphQL queries to include new fields
- Update UI components to use new fields with appropriate fallbacks

#### 6. Verification Checklist
Before completing GraphQL-related changes:
- [ ] `task graphql` executed successfully
- [ ] `go vet ./...` passes without GraphQL field errors
- [ ] `go test ./...` passes (especially mock-related tests)
- [ ] Frontend GraphQL queries include new fields
- [ ] Server restart planned (GraphQL schema cache)
- [ ] Browser cache clear planned (client-side GraphQL cache)

#### 7. Common Pitfalls to Avoid
- **Never commit schema changes without regenerating GraphQL code**
- **Never modify generated GraphQL files manually**
- **Always update test mocks when adding interface methods**
- **Remember that both frontend and backend caches need clearing after schema changes**
- **Check that all constructor signatures are updated consistently**

#### 8. Deployment Notes
- GraphQL schema changes require server restart to take effect
- Consider client cache invalidation strategies
- Test with real data to ensure proper field resolution

## Common Development Commands

### Building and Testing
- `go build` - Build the main binary
- `go test ./...` - Run all tests
- `go test ./pkg/path/to/package` - Run tests for specific package
- `task` - Run default tasks (mock generation)
- `task mock` (alias: `task m`) - Generate all mock files

### Code Generation
- `go install github.com/matryer/moq@latest` - Install moq for mock generation
- `moq -out pkg/domain/mock/interfaces.go ./pkg/domain/interfaces SlackClient` - Generate mocks

### GraphQL Development
- `task graphql` (alias: `task g`) - Regenerate GraphQL schema and resolvers
- `task` - Run all code generation (includes GraphQL and mocks)
- **Always run after modifying `graphql/schema.graphql`**
- **Always verify generated code compiles before committing**

## Important Development Guidelines

### Error Handling
- Use `github.com/m-mizutani/goerr/v2` for error handling
- Must wrap errors with `goerr.Wrap` to maintain error context
- Add helpful variables with `goerr.V` for debugging (e.g., `goerr.V("key", value)`)

### Performance
- Avoid premature optimization
- Implement simple solutions first, optimize only when actual performance issues are identified
- Do not add caching or complex optimizations without proven need

### Testing
- Use `github.com/m-mizutani/gt` package for type-safe testing
- Test files should have `package {name}_test`. Do not use same package name
- **Test MUST be included in same name test file. (e.g. test for `abc.go` must be in `abc_test.go`)**
- Use mock implementations from `pkg/domain/mock` generated by moq

### Code Visibility
- Do not expose unnecessary methods, variables and types
- Use `export_test.go` to expose items needed only for testing

## Architecture

### Core Structure
The application follows Domain-Driven Design (DDD) with clean architecture:

- `pkg/domain/` - Domain layer with business logic, interfaces, and models
- `pkg/service/` - Application services implementing business operations
- `pkg/controller/` - Interface adapters (HTTP, Slack)
- `pkg/usecase/` - Application use cases orchestrating domain operations
- `pkg/repository/` - Data persistence layer (Firestore, Memory)
- `pkg/cli/` - CLI command processing

### Key Components

#### Slack Integration
- `/hooks/slack/events` - Slack event webhook endpoint
- `/hooks/slack/interaction` - Slack interaction endpoint (future)
- Thread-based message replies
- Signature verification for security

#### Path Structure
```
/hooks/               # External service webhooks
  /slack/
    /events           # Slack events
    /interaction      # Slack interactions (future)
/api/                 # REST API (future)
/graphql              # GraphQL (future)
/                     # Frontend static files (future)
```

### Key Interfaces
- `interfaces.SlackClient` - Slack API client abstraction
- `interfaces.SlackEventUseCases` - Slack event handling use cases
- `interfaces.ThreadRepository` - Thread and message persistence abstraction

## Configuration

The application is configured via CLI flags or environment variables:
- Slack OAuth token (`--slack-oauth-token` or `TAMAMO_SLACK_OAUTH_TOKEN`)
- Slack signing secret (`--slack-signing-secret` or `TAMAMO_SLACK_SIGNING_SECRET`)
- Server address (`--addr` or `TAMAMO_ADDR`)
- Firestore Project ID (`--firestore-project-id` or `TAMAMO_FIRESTORE_PROJECT_ID`)
- Firestore Database ID (`--firestore-database-id` or `TAMAMO_FIRESTORE_DATABASE_ID`)

### Repository Selection
- If Firestore Project ID is provided, Firestore will be used for persistence
- Otherwise, in-memory repository will be used (data will be lost on restart)
- Firestore uses Application Default Credentials (ADC) for authentication

## Testing

Test files follow Go conventions (`*_test.go`). The codebase includes:
- Unit tests for individual components
- Integration tests with mock dependencies
- Mock generation using `moq` tool managed by Taskfile
- Common test suite for repository implementations

### Testing Repository Implementations
For Firestore tests, set the following environment variables:
```bash
export TEST_FIRESTORE_PROJECT="your-test-project"
export TEST_FIRESTORE_DATABASE="(default)"
```

