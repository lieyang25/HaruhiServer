package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"HaruhiServer/internal/repository/memory"
	"HaruhiServer/internal/service"
	transporthttp "HaruhiServer/internal/transport/http"
)

type capturedRecord struct {
	level slog.Level
	msg   string
	attrs map[string]any
}

type captureHandler struct {
	mu      sync.Mutex
	records []capturedRecord
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.mu.Lock()
	h.records = append(h.records, capturedRecord{level: r.Level, msg: r.Message, attrs: attrs})
	h.mu.Unlock()
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func (h *captureHandler) snapshot() []capturedRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]capturedRecord, len(h.records))
	copy(out, h.records)
	return out
}

func loggerWithCapture() (*slog.Logger, *captureHandler) {
	cap := &captureHandler{}
	return slog.New(cap), cap
}

type responseEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeEnvelope(t *testing.T, body string) responseEnvelope {
	t.Helper()
	var env responseEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", body, err)
	}
	return env
}

func decodeEnvelopeData(t *testing.T, body string) map[string]any {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", body, err)
	}
	data, _ := raw["data"].(map[string]any)
	if data == nil {
		t.Fatalf("data is missing in response: %q", body)
	}
	return data
}

func TestHealthzRoutes(t *testing.T) {
	logger, _ := loggerWithCapture()
	h := transporthttp.NewHandler(logger)

	getReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", getRec.Code, http.StatusOK)
	}
	if env := decodeEnvelope(t, getRec.Body.String()); env.Code != "OK" {
		t.Fatalf("GET /healthz code = %q, want OK", env.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /healthz status = %d, want %d", postRec.Code, http.StatusMethodNotAllowed)
	}
	if env := decodeEnvelope(t, postRec.Body.String()); env.Code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("POST /healthz code = %q, want METHOD_NOT_ALLOWED", env.Code)
	}
}

func TestReadyzRoutes(t *testing.T) {
	logger, _ := loggerWithCapture()
	h := transporthttp.NewHandler(logger)

	getReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /readyz status = %d, want %d", getRec.Code, http.StatusOK)
	}
	if env := decodeEnvelope(t, getRec.Body.String()); env.Code != "OK" {
		t.Fatalf("GET /readyz code = %q, want OK", env.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/readyz", nil)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /readyz status = %d, want %d", postRec.Code, http.StatusMethodNotAllowed)
	}
	if env := decodeEnvelope(t, postRec.Body.String()); env.Code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("POST /readyz code = %q, want METHOD_NOT_ALLOWED", env.Code)
	}
}

func TestSystemInfoRoute(t *testing.T) {
	logger, _ := loggerWithCapture()
	startedAt := time.Unix(1710000000, 0).UTC()
	h := transporthttp.NewHandlerWithSystemInfo(logger, transporthttp.SystemInfo{
		Name:      "haruhiserver",
		Version:   "v0.0.1",
		StartedAt: startedAt,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/info", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/system/info status = %d, want %d", rec.Code, http.StatusOK)
	}

	data := decodeEnvelopeData(t, rec.Body.String())
	if got, _ := data["service"].(string); got != "haruhiserver" {
		t.Fatalf("service = %q, want haruhiserver", got)
	}
	if got, _ := data["version"].(string); got != "v0.0.1" {
		t.Fatalf("version = %q, want v0.0.1", got)
	}
	if got, _ := data["started_at"].(string); got != startedAt.Format(time.RFC3339) {
		t.Fatalf("started_at = %q, want %q", got, startedAt.Format(time.RFC3339))
	}
	if got, ok := data["uptime_seconds"].(float64); !ok || got < 0 {
		t.Fatalf("uptime_seconds = %#v, want non-negative number", data["uptime_seconds"])
	}
}

func TestSystemInfoRoute_MethodNotAllowed(t *testing.T) {
	logger, _ := loggerWithCapture()
	h := transporthttp.NewHandler(logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/info", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/v1/system/info status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if env := decodeEnvelope(t, rec.Body.String()); env.Code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("POST /api/v1/system/info code = %q, want METHOD_NOT_ALLOWED", env.Code)
	}
}

func TestRequestIDMiddleware_UsesProvidedHeader(t *testing.T) {
	h := transporthttp.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), transporthttp.RequestIDMiddleware())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(transporthttp.RequestIDHeader, "req-123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get(transporthttp.RequestIDHeader); got != "req-123" {
		t.Fatalf("%s = %q, want req-123", transporthttp.RequestIDHeader, got)
	}
}

func TestCORSMiddleware_OPTIONS(t *testing.T) {
	h := transporthttp.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not run for OPTIONS")
	}), transporthttp.CORSMiddleware())

	req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
}

func TestChain_Order(t *testing.T) {
	calls := make([]string, 0, 4)
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "mw1-before")
			next.ServeHTTP(w, r)
			calls = append(calls, "mw1-after")
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "mw2-before")
			next.ServeHTTP(w, r)
			calls = append(calls, "mw2-after")
		})
	}

	h := transporthttp.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls = append(calls, "handler")
	}), mw1, mw2)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(calls) != len(want) {
		t.Fatalf("calls len = %d, want %d; calls=%v", len(calls), len(want), calls)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls[%d] = %q, want %q; calls=%v", i, calls[i], want[i], calls)
		}
	}
}

func TestRecoverMiddleware(t *testing.T) {
	logger, _ := loggerWithCapture()
	h := transporthttp.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), transporthttp.RecoverMiddleware(logger))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if env := decodeEnvelope(t, rec.Body.String()); env.Code != "INTERNAL" {
		t.Fatalf("code = %q, want INTERNAL", env.Code)
	}
}

func TestProjectCRUDRoutes(t *testing.T) {
	logger, _ := loggerWithCapture()
	repos := memory.NewRepositories()
	svcs, err := service.NewServices(repos, service.RandomIDGenerator{}, time.Now)
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	owner, err := svcs.Users.Create(context.Background(), service.CreateUserInput{
		Username:    "owner-http",
		DisplayName: "Owner HTTP",
		Email:       "owner-http@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}

	h := transporthttp.NewHandlerWithServices(logger, transporthttp.SystemInfo{}, svcs)

	createBody := []byte(`{"owner_id":"` + string(owner.ID) + `","name":"p9","description":"demo","visibility":"private"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(createBody))
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/projects status = %d, want %d body=%s", createRec.Code, http.StatusOK, createRec.Body.String())
	}
	createData := decodeEnvelopeData(t, createRec.Body.String())
	projectID, _ := createData["ID"].(string)
	if projectID == "" {
		t.Fatalf("POST /api/v1/projects missing ID: %s", createRec.Body.String())
	}

	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID, nil))
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/projects/{id} status = %d, want %d", getRec.Code, http.StatusOK)
	}

	patchRec := httptest.NewRecorder()
	patchBody := []byte(`{"name":"p9-updated","archived":true}`)
	h.ServeHTTP(patchRec, httptest.NewRequest(http.MethodPatch, "/api/v1/projects/"+projectID, bytes.NewReader(patchBody)))
	if patchRec.Code != http.StatusOK {
		t.Fatalf("PATCH /api/v1/projects/{id} status = %d, want %d body=%s", patchRec.Code, http.StatusOK, patchRec.Body.String())
	}

	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/projects status = %d, want %d", listRec.Code, http.StatusOK)
	}

	deleteRec := httptest.NewRecorder()
	h.ServeHTTP(deleteRec, httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+projectID, nil))
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("DELETE /api/v1/projects/{id} status = %d, want %d", deleteRec.Code, http.StatusOK)
	}
}

func TestProjectTaskRoutes_CRUDByProject(t *testing.T) {
	logger, _ := loggerWithCapture()
	repos := memory.NewRepositories()
	svcs, err := service.NewServices(repos, service.RandomIDGenerator{}, time.Now)
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	owner, err := svcs.Users.Create(context.Background(), service.CreateUserInput{
		Username:    "owner-task-http",
		DisplayName: "Owner Task HTTP",
		Email:       "owner-task-http@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create(owner) err = %v", err)
	}
	assignee, err := svcs.Users.Create(context.Background(), service.CreateUserInput{
		Username:    "assignee-task-http",
		DisplayName: "Assignee Task HTTP",
		Email:       "assignee-task-http@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create(assignee) err = %v", err)
	}
	project, err := svcs.Projects.Create(context.Background(), service.CreateProjectInput{
		OwnerID: owner.ID,
		Name:    "p10-project",
	})
	if err != nil {
		t.Fatalf("Projects.Create err = %v", err)
	}

	h := transporthttp.NewHandlerWithServices(logger, transporthttp.SystemInfo{}, svcs)

	dueAt := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	createBody := []byte(`{"creator_id":"` + string(owner.ID) + `","assignee_id":"` + string(assignee.ID) + `","title":"p10-task","description":"demo","priority":"high","due_at":"` + dueAt + `"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+string(project.ID)+"/tasks", bytes.NewReader(createBody))
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/projects/{id}/tasks status = %d, want %d body=%s", createRec.Code, http.StatusOK, createRec.Body.String())
	}
	createData := decodeEnvelopeData(t, createRec.Body.String())
	taskID, _ := createData["ID"].(string)
	if taskID == "" {
		t.Fatalf("POST /api/v1/projects/{id}/tasks missing ID: %s", createRec.Body.String())
	}

	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+string(project.ID)+"/tasks", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/projects/{id}/tasks status = %d, want %d body=%s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	listData := decodeEnvelopeData(t, listRec.Body.String())
	items, ok := listData["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("GET /api/v1/projects/{id}/tasks items = %#v, want len=1", listData["items"])
	}

	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+string(project.ID)+"/tasks/"+taskID, nil))
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/projects/{id}/tasks/{taskId} status = %d, want %d body=%s", getRec.Code, http.StatusOK, getRec.Body.String())
	}

	patchBody := []byte(`{"title":"p10-task-updated","description":"updated","priority":"urgent","clear_assignee":true,"clear_due_at":true}`)
	patchRec := httptest.NewRecorder()
	h.ServeHTTP(patchRec, httptest.NewRequest(http.MethodPatch, "/api/v1/projects/"+string(project.ID)+"/tasks/"+taskID, bytes.NewReader(patchBody)))
	if patchRec.Code != http.StatusOK {
		t.Fatalf("PATCH /api/v1/projects/{id}/tasks/{taskId} status = %d, want %d body=%s", patchRec.Code, http.StatusOK, patchRec.Body.String())
	}
	patchData := decodeEnvelopeData(t, patchRec.Body.String())
	if got, _ := patchData["Title"].(string); got != "p10-task-updated" {
		t.Fatalf("patched title = %q, want p10-task-updated", got)
	}
	if patchData["AssigneeID"] != nil {
		t.Fatalf("patched assignee = %#v, want nil", patchData["AssigneeID"])
	}
	if patchData["DueAt"] != nil {
		t.Fatalf("patched due_at = %#v, want nil", patchData["DueAt"])
	}

	deleteRec := httptest.NewRecorder()
	h.ServeHTTP(deleteRec, httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+string(project.ID)+"/tasks/"+taskID, nil))
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("DELETE /api/v1/projects/{id}/tasks/{taskId} status = %d, want %d body=%s", deleteRec.Code, http.StatusOK, deleteRec.Body.String())
	}
	deleteData := decodeEnvelopeData(t, deleteRec.Body.String())
	if deleted, _ := deleteData["deleted"].(bool); !deleted {
		t.Fatalf("delete result = %#v, want deleted=true", deleteData["deleted"])
	}

	getAfterDeleteRec := httptest.NewRecorder()
	h.ServeHTTP(getAfterDeleteRec, httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+string(project.ID)+"/tasks/"+taskID, nil))
	if getAfterDeleteRec.Code != http.StatusNotFound {
		t.Fatalf("GET deleted task status = %d, want %d body=%s", getAfterDeleteRec.Code, http.StatusNotFound, getAfterDeleteRec.Body.String())
	}
	if env := decodeEnvelope(t, getAfterDeleteRec.Body.String()); env.Code != "NOT_FOUND" {
		t.Fatalf("GET deleted task code = %q, want NOT_FOUND", env.Code)
	}
}

func TestProjectTaskRoutes_ProjectMismatchReturnsNotFound(t *testing.T) {
	logger, _ := loggerWithCapture()
	repos := memory.NewRepositories()
	svcs, err := service.NewServices(repos, service.RandomIDGenerator{}, time.Now)
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	owner, err := svcs.Users.Create(context.Background(), service.CreateUserInput{
		Username:    "owner-mismatch",
		DisplayName: "Owner Mismatch",
		Email:       "owner-mismatch@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}
	projectA, err := svcs.Projects.Create(context.Background(), service.CreateProjectInput{
		OwnerID: owner.ID,
		Name:    "project-a",
	})
	if err != nil {
		t.Fatalf("Projects.Create(projectA) err = %v", err)
	}
	projectB, err := svcs.Projects.Create(context.Background(), service.CreateProjectInput{
		OwnerID: owner.ID,
		Name:    "project-b",
	})
	if err != nil {
		t.Fatalf("Projects.Create(projectB) err = %v", err)
	}
	task, err := svcs.Tasks.Create(context.Background(), service.CreateTaskInput{
		ProjectID: projectA.ID,
		CreatorID: owner.ID,
		Title:     "mismatch-task",
	})
	if err != nil {
		t.Fatalf("Tasks.Create err = %v", err)
	}

	h := transporthttp.NewHandlerWithServices(logger, transporthttp.SystemInfo{}, svcs)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+string(projectB.ID)+"/tasks/"+string(task.ID), nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-project GET status = %d, want %d body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if env := decodeEnvelope(t, rec.Body.String()); env.Code != "NOT_FOUND" {
		t.Fatalf("cross-project GET code = %q, want NOT_FOUND", env.Code)
	}
}

func TestProjectTaskRoutes_InvalidDueAtReturnsInvalidArgument(t *testing.T) {
	logger, _ := loggerWithCapture()
	repos := memory.NewRepositories()
	svcs, err := service.NewServices(repos, service.RandomIDGenerator{}, time.Now)
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	owner, err := svcs.Users.Create(context.Background(), service.CreateUserInput{
		Username:    "owner-invalid-due",
		DisplayName: "Owner Invalid Due",
		Email:       "owner-invalid-due@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}
	project, err := svcs.Projects.Create(context.Background(), service.CreateProjectInput{
		OwnerID: owner.ID,
		Name:    "project-invalid-due",
	})
	if err != nil {
		t.Fatalf("Projects.Create err = %v", err)
	}

	h := transporthttp.NewHandlerWithServices(logger, transporthttp.SystemInfo{}, svcs)

	createBody := []byte(`{"creator_id":"` + string(owner.ID) + `","title":"invalid-due","due_at":"not-rfc3339"}`)
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+string(project.ID)+"/tasks", bytes.NewReader(createBody)))
	if createRec.Code != http.StatusBadRequest {
		t.Fatalf("POST with invalid due_at status = %d, want %d body=%s", createRec.Code, http.StatusBadRequest, createRec.Body.String())
	}
	if env := decodeEnvelope(t, createRec.Body.String()); env.Code != "INVALID_ARGUMENT" {
		t.Fatalf("POST with invalid due_at code = %q, want INVALID_ARGUMENT", env.Code)
	}
}

func TestRisk_RequestIDShouldBePresentInAccessLogs(t *testing.T) {
	logger, cap := loggerWithCapture()
	h := transporthttp.NewHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	for _, log := range cap.snapshot() {
		if log.msg == "http request" {
			rid, _ := log.attrs["request_id"].(string)
			if rid == "" {
				t.Fatalf("semantic risk: request_id should be non-empty in access log")
			}
			return
		}
	}
	t.Fatal("no http request log captured")
}

func TestRisk_RequestIDShouldBePresentInPanicLogs(t *testing.T) {
	logger, cap := loggerWithCapture()
	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	h := transporthttp.Chain(
		mux,
		transporthttp.CORSMiddleware(),
		transporthttp.RecoverMiddleware(logger),
		transporthttp.LoggingMiddleware(logger),
		transporthttp.RequestIDMiddleware(),
	)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	for _, log := range cap.snapshot() {
		if log.msg == "panic recovered" {
			rid, _ := log.attrs["request_id"].(string)
			if rid == "" {
				t.Fatalf("semantic risk: panic log should include non-empty request_id")
			}
			return
		}
	}
	t.Fatal("no panic recovered log captured")
}

func TestRisk_OPTIONSShouldCarryRequestIDAndBeLogged(t *testing.T) {
	logger, cap := loggerWithCapture()
	h := transporthttp.NewHandler(logger)

	req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get(transporthttp.RequestIDHeader) == "" {
		t.Fatalf("semantic risk: OPTIONS response should include %s", transporthttp.RequestIDHeader)
	}
	for _, log := range cap.snapshot() {
		if log.msg == "http request" {
			return
		}
	}
	t.Fatalf("semantic risk: OPTIONS should be observable in access logs")
}

type countingResponseWriter struct {
	header http.Header
	calls  int
}

func newCountingResponseWriter() *countingResponseWriter {
	return &countingResponseWriter{header: make(http.Header)}
}

func (w *countingResponseWriter) Header() http.Header         { return w.header }
func (w *countingResponseWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *countingResponseWriter) WriteHeader(int)             { w.calls++ }

func TestRisk_LoggingMiddlewareShouldNotForwardDuplicateWriteHeader(t *testing.T) {
	logger, _ := loggerWithCapture()
	h := transporthttp.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.WriteHeader(http.StatusAccepted)
	}), transporthttp.LoggingMiddleware(logger))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	base := newCountingResponseWriter()
	h.ServeHTTP(base, req)

	if base.calls != 1 {
		t.Fatalf("semantic risk: underlying WriteHeader called %d times, want 1", base.calls)
	}
}
