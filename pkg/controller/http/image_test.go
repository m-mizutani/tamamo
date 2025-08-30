package http_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/gt"
	httpCtrl "github.com/m-mizutani/tamamo/pkg/controller/http"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	imageService "github.com/m-mizutani/tamamo/pkg/service/image"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

// mockAgentUseCase implements interfaces.AgentUseCases for testing
type mockAgentUseCase struct {
	agents map[string]*agent.Agent
}

func newMockAgentUseCase() *mockAgentUseCase {
	return &mockAgentUseCase{
		agents: make(map[string]*agent.Agent),
	}
}

func (m *mockAgentUseCase) GetAgent(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
	if agent, exists := m.agents[id.String()]; exists {
		return &interfaces.AgentWithVersion{Agent: agent}, nil
	}
	return nil, fmt.Errorf("agent not found")
}

func (m *mockAgentUseCase) addAgent(agent *agent.Agent) {
	m.agents[agent.ID.String()] = agent
}

// Implement other required methods (not used in tests)
func (m *mockAgentUseCase) CreateAgent(ctx context.Context, req *interfaces.CreateAgentRequest) (*agent.Agent, error) {
	return nil, nil
}
func (m *mockAgentUseCase) UpdateAgent(ctx context.Context, id types.UUID, req *interfaces.UpdateAgentRequest) (*agent.Agent, error) {
	return nil, nil
}
func (m *mockAgentUseCase) DeleteAgent(ctx context.Context, id types.UUID) error { return nil }
func (m *mockAgentUseCase) ArchiveAgent(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
	return nil, nil
}
func (m *mockAgentUseCase) UnarchiveAgent(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
	return nil, nil
}
func (m *mockAgentUseCase) CreateAgentVersion(ctx context.Context, req *interfaces.CreateVersionRequest) (*agent.AgentVersion, error) {
	return nil, nil
}
func (m *mockAgentUseCase) ListAgents(ctx context.Context, offset, limit int) (*interfaces.AgentListResponse, error) {
	return nil, nil
}
func (m *mockAgentUseCase) ListAllAgents(ctx context.Context, offset, limit int) (*interfaces.AgentListResponse, error) {
	return nil, nil
}
func (m *mockAgentUseCase) ListAgentsByStatus(ctx context.Context, status agent.Status, offset, limit int) (*interfaces.AgentListResponse, error) {
	return nil, nil
}
func (m *mockAgentUseCase) CheckAgentIDAvailability(ctx context.Context, agentID string) (*interfaces.AgentIDAvailability, error) {
	return nil, nil
}
func (m *mockAgentUseCase) GetAgentVersions(ctx context.Context, agentUUID types.UUID) ([]*agent.AgentVersion, error) {
	return nil, nil
}
func (m *mockAgentUseCase) ValidateAgentID(agentID string) error {
	return nil
}
func (m *mockAgentUseCase) ValidateVersion(version string) error {
	return nil
}

// mockStorageAdapter for testing
type mockStorageAdapter struct {
	storage map[string][]byte
}

func newMockStorageAdapter() *mockStorageAdapter {
	return &mockStorageAdapter{
		storage: make(map[string][]byte),
	}
}

func (m *mockStorageAdapter) Put(ctx context.Context, key string, data []byte) error {
	m.storage[key] = data
	return nil
}

func (m *mockStorageAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	data, exists := m.storage[key]
	if !exists {
		return nil, interfaces.ErrStorageKeyNotFound
	}
	return data, nil
}

// createTestJPEG creates a test JPEG image
func createTestJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestImageController_HandleUploadAgentImage(t *testing.T) {
	ctx := context.Background()

	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	agentUseCase := newMockAgentUseCase()
	config := imageService.DefaultProcessorConfig()
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)
	imageUseCase := usecase.NewImageUseCases(processor, repository, agentUseCase)
	controller := httpCtrl.NewImageController(imageUseCase)

	// Create test agent
	agentID := types.NewUUID(ctx)
	testAgent := &agent.Agent{
		ID:      agentID,
		AgentID: "test-agent",
		Name:    "Test Agent",
	}
	agentUseCase.addAgent(testAgent)

	// Create test image
	jpegData := createTestJPEG()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create form file part with proper Content-Type header
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="test.jpg"`)
	h.Set("Content-Type", "image/jpeg")
	part, err := writer.CreatePart(h)
	gt.NoError(t, err).Required()
	_, err = part.Write(jpegData)
	gt.NoError(t, err).Required()
	writer.Close()

	// Create request
	req := httptest.NewRequest("POST", "/api/agents/"+agentID.String()+"/image", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create router with chi to handle URL parameters
	r := chi.NewRouter()
	r.Post("/api/agents/{agentID}/image", controller.HandleUploadAgentImage)

	// Record response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		// Debug: try to process the image directly to see the error
		_, directErr := processor.ProcessAndStore(ctx, agentID, bytes.NewReader(jpegData), "image/jpeg", int64(len(jpegData)))
		if directErr != nil {
			t.Errorf("Direct processor error: %v", directErr)
		}
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestImageController_HandleGetAgentImage(t *testing.T) {
	ctx := context.Background()

	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	agentUseCase := newMockAgentUseCase()
	config := imageService.DefaultProcessorConfig()
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)
	imageUseCase := usecase.NewImageUseCases(processor, repository, agentUseCase)
	controller := httpCtrl.NewImageController(imageUseCase)

	// Create test agent and image
	agentID := types.NewUUID(ctx)
	testAgent := &agent.Agent{
		ID:      agentID,
		AgentID: "test-agent",
		Name:    "Test Agent",
	}
	agentUseCase.addAgent(testAgent)

	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	// Process and store image
	agentImage, err := processor.ProcessAndStore(ctx, agentID, reader, "image/jpeg", int64(len(jpegData)))
	gt.NoError(t, err).Required()

	// Update agent with image ID
	testAgent.ImageID = &agentImage.ID
	agentUseCase.addAgent(testAgent)

	// Create request
	req := httptest.NewRequest("GET", "/api/agents/"+agentID.String()+"/image", nil)

	// Create router
	r := chi.NewRouter()
	r.Get("/api/agents/{agentID}/image", controller.HandleGetAgentImage)

	// Record response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("Expected Content-Type image/jpeg, got %s", w.Header().Get("Content-Type"))
	}

	// Check that we got image data
	if len(w.Body.Bytes()) == 0 {
		t.Error("Expected image data, got empty response")
	}
}

func TestImageController_HandleGetAgentImageInfo(t *testing.T) {
	ctx := context.Background()

	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	agentUseCase := newMockAgentUseCase()
	config := imageService.DefaultProcessorConfig()
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)
	imageUseCase := usecase.NewImageUseCases(processor, repository, agentUseCase)
	controller := httpCtrl.NewImageController(imageUseCase)

	// Create test agent and image
	agentID := types.NewUUID(ctx)
	testAgent := &agent.Agent{
		ID:      agentID,
		AgentID: "test-agent",
		Name:    "Test Agent",
	}
	agentUseCase.addAgent(testAgent)

	jpegData := createTestJPEG()
	reader := bytes.NewReader(jpegData)

	// Process and store image
	agentImage, err := processor.ProcessAndStore(ctx, agentID, reader, "image/jpeg", int64(len(jpegData)))
	gt.NoError(t, err).Required()

	// Update agent with image ID
	testAgent.ImageID = &agentImage.ID
	agentUseCase.addAgent(testAgent)

	// Create request
	req := httptest.NewRequest("GET", "/api/agents/"+agentID.String()+"/image/info", nil)

	// Create router
	r := chi.NewRouter()
	r.Get("/api/agents/{agentID}/image/info", controller.HandleGetAgentImageInfo)

	// Record response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestImageController_HandleUploadAgentImage_InvalidAgentID(t *testing.T) {
	// Setup
	validator := imageService.NewValidator()
	storage := newMockStorageAdapter()
	repository := memory.NewAgentImageRepository()
	agentUseCase := newMockAgentUseCase()
	config := imageService.DefaultProcessorConfig()
	processor := imageService.NewProcessor(validator, storage, repository, nil, config)
	imageUseCase := usecase.NewImageUseCases(processor, repository, agentUseCase)
	controller := httpCtrl.NewImageController(imageUseCase)

	// Create request with invalid agent ID
	req := httptest.NewRequest("POST", "/api/agents/invalid-id/image", nil)

	// Create router
	r := chi.NewRouter()
	r.Post("/api/agents/{agentID}/image", controller.HandleUploadAgentImage)

	// Record response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
