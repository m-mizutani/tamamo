package graphql_test

import (
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/controller/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
)

func TestNewResolver(t *testing.T) {
	mockRepo := &mock.ThreadRepositoryMock{}
	resolver := graphql.NewResolver(mockRepo)

	gt.V(t, resolver).NotNil()
}

func TestResolver_DependencyInjection(t *testing.T) {
	mockRepo := &mock.ThreadRepositoryMock{}
	resolver := graphql.NewResolver(mockRepo)

	// Verify that resolver can be created with mock repository
	gt.V(t, resolver).NotNil()

	// Verify that Query resolver can be obtained
	queryResolver := resolver.Query()
	gt.V(t, queryResolver).NotNil()

	// Verify that Thread resolver can be obtained
	threadResolver := resolver.Thread()
	gt.V(t, threadResolver).NotNil()
}
