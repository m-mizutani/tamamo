package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-mizutani/gt"
	graphql_controller "github.com/m-mizutani/tamamo/pkg/controller/graphql"
	server "github.com/m-mizutani/tamamo/pkg/controller/http"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   any `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

func TestServer_GraphQLEndpoint(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()
	agentRepo := memory.NewAgentMemoryClient()
	agentUseCase := usecase.NewAgentUseCases(agentRepo)

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, agentUseCase, nil, nil, nil, nil, nil, nil) // nil for user usecase, factory, image processor, image repo, and integrations for tests

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Test GraphQL endpoint exists
	req := httptest.NewRequest("POST", "/graphql", nil)
	rec := httptest.NewRecorder()

	httpServer.ServeHTTP(rec, req)

	// Should not return 404 (endpoint exists)
	gt.V(t, rec.Code).NotEqual(http.StatusNotFound)
}

func TestServer_GraphQLQuery_Threads(t *testing.T) {
	// Setup memory repository with test data
	memRepo := memory.New()

	// Create test threads
	ctx := context.Background()
	testThread1, err := memRepo.GetOrPutThread(ctx, "T123456", "C123456", "1234567890.123456")
	gt.NoError(t, err)
	testThread2, err := memRepo.GetOrPutThread(ctx, "T123456", "C123456", "1234567891.123456")
	gt.NoError(t, err)

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Prepare GraphQL query
	query := GraphQLRequest{
		Query: `{
			threads(offset: 0, limit: 10) {
				threads {
					id
					teamId
					channelId
					threadTs
					createdAt
					updatedAt
				}
				totalCount
			}
		}`,
	}

	// Marshal query to JSON
	queryBytes, err := json.Marshal(query)
	gt.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(queryBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	httpServer.ServeHTTP(rec, req)

	// Verify response
	gt.Equal(t, rec.Code, http.StatusOK)
	gt.V(t, rec.Header().Get("Content-Type")).Equal("application/json")

	// Parse response
	var response GraphQLResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	gt.NoError(t, err)

	// Verify GraphQL response structure
	gt.V(t, response.Data).NotNil()
	gt.Equal(t, len(response.Errors), 0)

	// Verify data content
	data := response.Data.(map[string]any)
	threadsData := data["threads"].(map[string]any)
	threads := threadsData["threads"].([]any)
	totalCount := threadsData["totalCount"].(float64)

	gt.Equal(t, len(threads), 2)
	gt.Equal(t, int(totalCount), 2)

	// Verify first thread data (order may vary due to sorting)
	thread1 := threads[0].(map[string]any)
	thread1ID := thread1["id"].(string)
	gt.V(t, thread1ID == string(testThread1.ID) || thread1ID == string(testThread2.ID)).Equal(true)
}

func TestServer_GraphQLQuery_Thread_Success(t *testing.T) {
	// Setup memory repository with test data
	memRepo := memory.New()

	// Create a test thread
	ctx := context.Background()
	testThread, err := memRepo.GetOrPutThread(ctx, "T123456", "C123456", "1234567890.123456")
	gt.NoError(t, err)

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Prepare GraphQL query for specific thread
	query := GraphQLRequest{
		Query: fmt.Sprintf(`{
			thread(id: "%s") {
				id
				teamId
				channelId
				threadTs
				createdAt
				updatedAt
			}
		}`, testThread.ID),
	}

	// Marshal query to JSON
	queryBytes, err := json.Marshal(query)
	gt.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(queryBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	httpServer.ServeHTTP(rec, req)

	// Verify response
	gt.Equal(t, rec.Code, http.StatusOK)
	gt.V(t, rec.Header().Get("Content-Type")).Equal("application/json")

	// Parse response
	var response GraphQLResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	gt.NoError(t, err)

	// Verify GraphQL response structure
	gt.V(t, response.Data).NotNil()
	gt.Equal(t, len(response.Errors), 0)

	// Verify thread data in response
	data := response.Data.(map[string]any)
	thread := data["thread"].(map[string]any)
	gt.Equal(t, thread["id"].(string), string(testThread.ID))
	gt.Equal(t, thread["teamId"].(string), testThread.TeamID)
	gt.Equal(t, thread["channelId"].(string), testThread.ChannelID)
	gt.Equal(t, thread["threadTs"].(string), testThread.ThreadTS)
}

func TestServer_GraphQLQuery_Thread_InvalidID(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Prepare GraphQL query with invalid ID
	query := GraphQLRequest{
		Query: `{
			thread(id: "invalid-id") {
				id
				teamId
				channelId
				threadTs
				createdAt
				updatedAt
			}
		}`,
	}

	// Marshal query to JSON
	queryBytes, err := json.Marshal(query)
	gt.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(queryBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	httpServer.ServeHTTP(rec, req)

	// Verify response
	gt.Equal(t, rec.Code, http.StatusOK)

	// Parse response
	var response GraphQLResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	gt.NoError(t, err)

	// Should have GraphQL errors for invalid ID
	gt.V(t, len(response.Errors) > 0).Equal(true)
}

func TestServer_GraphiQLEndpoint_Disabled(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server without GraphiQL enabled
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
		server.WithGraphiQL(false),
	)

	// Test GraphiQL endpoint should return 404 when disabled
	req := httptest.NewRequest("GET", "/graphiql", nil)
	rec := httptest.NewRecorder()

	httpServer.ServeHTTP(rec, req)

	// Should return 404 when GraphiQL is disabled
	gt.Equal(t, rec.Code, http.StatusNotFound)
}

func TestServer_GraphiQLEndpoint_Enabled(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphiQL enabled
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
		server.WithGraphiQL(true),
	)

	// Test GraphiQL endpoint should work when enabled
	req := httptest.NewRequest("GET", "/graphiql", nil)
	rec := httptest.NewRecorder()

	httpServer.ServeHTTP(rec, req)

	// Should return 200 when GraphiQL is enabled
	gt.Equal(t, rec.Code, http.StatusOK)
	gt.V(t, rec.Header().Get("Content-Type")).Equal("text/html; charset=UTF-8")
}

func TestServer_GraphQLQuery_Thread_NotFound(t *testing.T) {
	// Setup memory repository (empty)
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Create a valid thread ID that doesn't exist in repository
	ctx := context.Background()
	testThread, err := memRepo.GetOrPutThread(ctx, "T123456", "C123456", "1234567890.123456")
	gt.NoError(t, err)
	nonExistentID := string(testThread.ID) + "999"

	// Prepare GraphQL query with non-existent ID
	query := GraphQLRequest{
		Query: fmt.Sprintf(`{
			thread(id: "%s") {
				id
				teamId
				channelId
				threadTs
			}
		}`, nonExistentID),
	}

	// Marshal query to JSON
	queryBytes, err := json.Marshal(query)
	gt.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(queryBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	httpServer.ServeHTTP(rec, req)

	// Verify response
	gt.Equal(t, rec.Code, http.StatusOK)

	// Parse response
	var response GraphQLResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	gt.NoError(t, err)

	// Should have GraphQL errors for non-existent thread
	gt.V(t, len(response.Errors) > 0).Equal(true)
	// For GraphQL, data field can still be present with null values
	if response.Data != nil {
		data := response.Data.(map[string]any)
		gt.V(t, data["thread"]).Nil()
	}
}

func TestServer_GraphQLQuery_InvalidSyntax(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Prepare invalid GraphQL query
	query := GraphQLRequest{
		Query: `{
			invalidSyntax {
				id
			}
		`, // Missing closing brace
	}

	// Marshal query to JSON
	queryBytes, err := json.Marshal(query)
	gt.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(queryBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	httpServer.ServeHTTP(rec, req)

	// Verify response
	gt.Equal(t, rec.Code, http.StatusUnprocessableEntity)
}

func TestServer_GraphQLQuery_InvalidJSON(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Create HTTP request with invalid JSON
	invalidJSON := `{"query": "{ threads { id } }` // Missing closing brace
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	httpServer.ServeHTTP(rec, req)

	// Verify response - should handle invalid JSON gracefully
	gt.Equal(t, rec.Code, http.StatusBadRequest)
}

func TestServer_GraphQLQuery_EmptyQuery(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Prepare empty GraphQL query
	query := GraphQLRequest{
		Query: "",
	}

	// Marshal query to JSON
	queryBytes, err := json.Marshal(query)
	gt.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(queryBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	httpServer.ServeHTTP(rec, req)

	// Verify response
	gt.Equal(t, rec.Code, http.StatusUnprocessableEntity)
}

func TestServer_GraphQLQuery_UnsupportedHTTPMethod(t *testing.T) {
	// Setup memory repository
	memRepo := memory.New()

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Test GET method (should be unsupported for GraphQL mutations)
	req := httptest.NewRequest("GET", "/graphql", nil)
	rec := httptest.NewRecorder()

	httpServer.ServeHTTP(rec, req)

	// Verify response - GraphQL endpoint should handle GET for queries
	gt.V(t, rec.Code).NotEqual(http.StatusNotFound)
}

func TestServer_GraphQLQuery_ConcurrentRequests(t *testing.T) {
	// Setup memory repository with test data
	memRepo := memory.New()

	// Create multiple test threads
	ctx := context.Background()
	var testThreads []*slack.Thread
	for i := 0; i < 5; i++ {
		thread, err := memRepo.GetOrPutThread(ctx, fmt.Sprintf("T%d", i), fmt.Sprintf("C%d", i), fmt.Sprintf("123456789%d.123456", i))
		gt.NoError(t, err)
		testThreads = append(testThreads, thread)
	}

	// Create GraphQL controller
	graphqlCtrl := graphql_controller.NewResolver(memRepo, nil, nil, nil, nil, nil, nil, nil)

	// Create HTTP server with GraphQL controller
	httpServer := server.New(
		server.WithGraphQLController(graphqlCtrl),
	)

	// Execute concurrent requests
	const concurrentRequests = 10
	ch := make(chan error, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func(threadIndex int) {
			thread := testThreads[threadIndex%len(testThreads)]

			// Prepare GraphQL query
			query := GraphQLRequest{
				Query: fmt.Sprintf(`{
					thread(id: "%s") {
						id
						teamId
						channelId
						threadTs
					}
				}`, thread.ID),
			}

			// Marshal query to JSON
			queryBytes, err := json.Marshal(query)
			if err != nil {
				ch <- err
				return
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(queryBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute request
			httpServer.ServeHTTP(rec, req)

			// Verify response
			if rec.Code != http.StatusOK {
				ch <- fmt.Errorf("expected status 200, got %d", rec.Code)
				return
			}

			// Parse response
			var response GraphQLResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			if err != nil {
				ch <- err
				return
			}

			// Verify response structure
			if response.Data == nil {
				ch <- fmt.Errorf("response data is nil")
				return
			}

			if len(response.Errors) > 0 {
				ch <- fmt.Errorf("unexpected GraphQL errors: %v", response.Errors)
				return
			}

			ch <- nil
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrentRequests; i++ {
		err := <-ch
		gt.NoError(t, err)
	}
}
