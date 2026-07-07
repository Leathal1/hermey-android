package endpoints

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Leathal1/hermey-android/core/client"
	"github.com/Leathal1/hermey-android/core/models"
)

func newTestClient(t *testing.T) (*client.APIClient, *httptest.Server) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
		case "/api/auth/status":
			w.Write([]byte(`{"auth_enabled":false,"logged_in":false}`))
		case "/api/auth/login", "/api/auth/logout":
			w.WriteHeader(http.StatusOK)
		case "/api/sessions":
			w.Write([]byte(`{"sessions":[{"id":"s1","title":"Test"}]}`))
		case "/api/session":
			w.Write([]byte(`{"session":{"id":"s1","title":"Test"},"messages":[{"id":"m1","role":"user","content":"hi"}]}`))
		case "/api/session/new":
			w.Write([]byte(`{"id":"s2","title":"New"}`))
		case "/api/session/rename", "/api/session/delete", "/api/session/pin", "/api/session/archive",
			"/api/session/merge", "/api/session/compact":
			w.WriteHeader(http.StatusOK)
		case "/api/session/fork":
			w.Write([]byte(`{"new_session_id":"s3"}`))
		case "/api/session/export":
			w.Write([]byte(`{"exported":true}`))
		case "/api/session/import":
			w.Write([]byte(`{"session_id":"s4"}`))
		case "/api/session/search":
			w.Write([]byte(`{"sessions":[]}`))
		case "/api/chat/start":
			w.Write([]byte(`{"stream_id":"st1","session_id":"s1"}`))
		case "/api/chat/cancel":
			w.WriteHeader(http.StatusOK)
		case "/api/chat/steer":
			w.Write([]byte(`{"accepted":true,"stream_id":"st1"}`))
		case "/api/chat/history":
			w.Write([]byte(`{"messages":[{"id":"m1","role":"assistant","content":"hello"}]}`))
		case "/api/chat/feedback":
			w.WriteHeader(http.StatusOK)
		case "/api/chat/stream/status":
			w.WriteHeader(http.StatusOK)
		case "/api/list":
			w.Write([]byte(`[{"name":"README.md","path":"README.md","is_dir":false,"size":12}]`))
		case "/api/file":
			w.Write([]byte(`{"path":"README.md","content":"hello world"}`))
		case "/api/file/update", "/api/file/delete":
			w.WriteHeader(http.StatusOK)
		case "/api/upload":
			w.Write([]byte(`{"filename":"up.txt","path":"up.txt","size":4}`))
		case "/api/models":
			w.Write([]byte(`[{"id":"gpt-4o","name":"GPT-4o","provider":"openai"}]`))
		case "/api/providers":
			w.Write([]byte(`[{"id":"openai","name":"OpenAI","enabled":true}]`))
		case "/api/profiles":
			w.Write([]byte(`[{"name":"coder","active":true}]`))
		case "/api/settings":
			w.Write([]byte(`{"bot_name":"Hermes","version":"0.16.0"}`))
		case "/api/settings/update":
			w.Write([]byte(`{"bot_name":"Updated","version":"0.16.0"}`))
		case "/api/reasoning":
			w.Write([]byte(`{"display":"collapsed","effort":"medium"}`))
		case "/api/reasoning/update":
			w.Write([]byte(`{"display":"expanded","effort":"high"}`))
		case "/api/mcp":
			w.Write([]byte(`[{"name":"filesystem","command":"npx","enabled":true}]`))
		case "/api/mcp/tools":
			w.Write([]byte(`[{"name":"read_file","description":"Read a file"}]`))
		case "/api/tools":
			w.Write([]byte(`[{"name":"web_search","enabled":true}]`))
		case "/api/crons":
			w.Write([]byte(`[{"id":"c1","name":"nightly","schedule":"0 0 * * *"}]`))
		case "/api/skills":
			w.Write([]byte(`[{"name":"github","description":"GitHub skill"}]`))
		case "/api/memory":
			w.Write([]byte(`[{"content":"note","target":"memory"}]`))
		case "/api/projects":
			w.Write([]byte(`[{"id":"p1","name":"Hermdroid"}]`))
		case "/api/kanban":
			w.Write([]byte(`[{"id":"t1","title":"Build core","status":"running"}]`))
		case "/api/inbox":
			w.Write([]byte(`[{"id":"i1","title":"Alert","body":"hello","read":false}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	c := client.NewAPIClientWithHTTP(ts.URL, ts.Client())
	c.SetBaseURL(ts.URL)
	return c, ts
}

func assertOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHealth(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	assertOK(t, Health(c))
}

func TestAuthStatus(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	status, err := GetAuthStatus(c)
	assertOK(t, err)
	if status.AuthEnabled {
		t.Error("expected auth_enabled=false")
	}
}

func TestLogin(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	assertOK(t, Login(c, &LoginRequest{Password: "secret"}))
}

func TestLogout(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	assertOK(t, Logout(c))
}

func TestListSessions(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	r, err := ListSessions(c)
	assertOK(t, err)
	if len(r.Sessions) != 1 || r.Sessions[0].ID != "s1" {
		t.Fatalf("unexpected sessions: %+v", r.Sessions)
	}
}

func TestGetSession(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	r, err := GetSession(c, &GetSessionRequest{SessionID: "s1", Messages: true})
	assertOK(t, err)
	if r.Session.ID != "s1" {
		t.Errorf("session id = %q, want s1", r.Session.ID)
	}
	if len(r.Messages) != 1 {
		t.Errorf("messages len = %d, want 1", len(r.Messages))
	}
}

func TestNewSession(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	s, err := NewSession(c, &NewSessionRequest{Model: "gpt-4o"})
	assertOK(t, err)
	if s.ID != "s2" {
		t.Errorf("id = %q, want s2", s.ID)
	}
}

func TestRenameDeletePinArchive(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	assertOK(t, RenameSession(c, &RenameSessionRequest{SessionID: "s1", Title: "Renamed"}))
	assertOK(t, DeleteSession(c, &DeleteSessionRequest{SessionID: "s1"}))
	assertOK(t, PinSession(c, &PinSessionRequest{SessionID: "s1", Pinned: true}))
	assertOK(t, ArchiveSession(c, &ArchiveSessionRequest{SessionID: "s1", Archived: true}))
}

func TestForkMergeCompact(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	r, err := ForkSession(c, &ForkSessionRequest{SessionID: "s1"})
	assertOK(t, err)
	if r.NewSessionID != "s3" {
		t.Errorf("new session id = %q, want s3", r.NewSessionID)
	}
	assertOK(t, MergeSession(c, &MergeSessionRequest{SourceID: "s1", TargetID: "s2"}))
	assertOK(t, CompactSession(c, &CompactSessionRequest{SessionID: "s1"}))
}

func TestExportImport(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	data, err := ExportSession(c, &ExportSessionRequest{SessionID: "s1", Format: "json"})
	assertOK(t, err)
	if !strings.Contains(string(data), "exported") {
		t.Errorf("expected exported payload, got %s", data)
	}
	ir, err := ImportSession(c, &ImportSessionRequest{Data: data, Title: "Imported"})
	assertOK(t, err)
	if ir.SessionID != "s4" {
		t.Errorf("session id = %q, want s4", ir.SessionID)
	}
}

func TestSearchSessions(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	r, err := SearchSessions(c, &SearchSessionsRequest{Query: "test"})
	assertOK(t, err)
	if r.Sessions == nil {
		t.Error("expected non-nil sessions slice")
	}
}

func TestStartCancelSteerChat(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	r, err := StartChat(c, &StartChatRequest{SessionID: "s1", Message: "hello"})
	assertOK(t, err)
	if r.StreamID != "st1" {
		t.Errorf("stream id = %q, want st1", r.StreamID)
	}
	assertOK(t, CancelChat(c, &CancelChatRequest{StreamID: "st1"}))
	sr, err := SteerChat(c, &SteerChatRequest{SessionID: "s1", Text: "go on"})
	assertOK(t, err)
	if !sr.Accepted {
		t.Error("expected accepted=true")
	}
}

func TestGetChatHistory(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	r, err := GetChatHistory(c, &GetChatHistoryRequest{SessionID: "s1"})
	assertOK(t, err)
	if len(r.Messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(r.Messages))
	}
	if r.Messages[0].Role != "assistant" {
		t.Errorf("role = %q, want assistant", r.Messages[0].Role)
	}
}

func TestSendFeedback(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	assertOK(t, SendFeedback(c, &SendFeedbackRequest{MessageID: "m1", Rating: "up"}))
}

func TestStreamStatus(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	ok, err := StreamStatus(c, &StreamStatusRequest{StreamID: "st1"})
	assertOK(t, err)
	if !ok {
		t.Error("expected stream active")
	}
}

func TestWorkspaceFileOps(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	entries, err := ListWorkspace(c, &ListWorkspaceRequest{SessionID: "s1"})
	assertOK(t, err)
	if len(entries) != 1 || entries[0].Name != "README.md" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
	fc, err := GetFile(c, &GetFileRequest{SessionID: "s1", Path: "README.md"})
	assertOK(t, err)
	if fc.Content != "hello world" {
		t.Errorf("content = %q, want hello world", fc.Content)
	}
	assertOK(t, UpdateFile(c, &UpdateFileRequest{SessionID: "s1", Path: "README.md", Content: "updated"}))
	assertOK(t, DeleteFile(c, &DeleteFileRequest{SessionID: "s1", Path: "README.md"}))
}

func TestUploadFile(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	ur, err := UploadFile(c, &UploadFileRequest{SessionID: "s1", Filename: "up.txt", Content: []byte("test")})
	assertOK(t, err)
	if ur.Filename != "up.txt" {
		t.Errorf("filename = %q, want up.txt", ur.Filename)
	}
}

func TestModelsProvidersProfiles(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	models, err := ListModels(c)
	assertOK(t, err)
	if len(models) != 1 || models[0].ID != "gpt-4o" {
		t.Fatalf("unexpected models: %+v", models)
	}
	providers, err := ListProviders(c)
	assertOK(t, err)
	if len(providers) != 1 || !providers[0].Enabled {
		t.Fatalf("unexpected providers: %+v", providers)
	}
	profiles, err := ListProfiles(c)
	assertOK(t, err)
	if len(profiles) != 1 || profiles[0].Name != "coder" {
		t.Fatalf("unexpected profiles: %+v", profiles)
	}
}

func TestSettings(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	s, err := GetSettings(c)
	assertOK(t, err)
	if s.BotName != "Hermes" {
		t.Errorf("bot_name = %q, want Hermes", s.BotName)
	}
	us, err := UpdateSettings(c, &UpdateSettingsRequest{BotName: "Updated"})
	assertOK(t, err)
	if us.BotName != "Updated" {
		t.Errorf("bot_name = %q, want Updated", us.BotName)
	}
}

func TestReasoning(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	r, err := GetReasoning(c)
	assertOK(t, err)
	if r.Display != "collapsed" {
		t.Errorf("display = %q, want collapsed", r.Display)
	}
	ur, err := UpdateReasoning(c, &UpdateReasoningRequest{Display: "expanded", Effort: "high"})
	assertOK(t, err)
	if ur.Display != "expanded" {
		t.Errorf("display = %q, want expanded", ur.Display)
	}
}

func TestMCPTools(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	configs, err := ListMCPConfigs(c)
	assertOK(t, err)
	if len(configs) != 1 || configs[0].Name != "filesystem" {
		t.Fatalf("unexpected mcp configs: %+v", configs)
	}
	tools, err := ListMCPTools(c)
	assertOK(t, err)
	if len(tools) != 1 || tools[0].Name != "read_file" {
		t.Fatalf("unexpected mcp tools: %+v", tools)
	}
	tcfgs, err := ListTools(c)
	assertOK(t, err)
	if len(tcfgs) != 1 || tcfgs[0].Name != "web_search" {
		t.Fatalf("unexpected tools: %+v", tcfgs)
	}
}

func TestReadOnlyPanels(t *testing.T) {
	c, ts := newTestClient(t)
	defer ts.Close()
	crons, err := ListCronJobs(c)
	assertOK(t, err)
	if len(crons) != 1 || crons[0].ID != "c1" {
		t.Fatalf("unexpected crons: %+v", crons)
	}
	skills, err := ListSkills(c)
	assertOK(t, err)
	if len(skills) != 1 || skills[0].Name != "github" {
		t.Fatalf("unexpected skills: %+v", skills)
	}
	memory, err := GetMemory(c)
	assertOK(t, err)
	if len(memory) != 1 || memory[0].Target != "memory" {
		t.Fatalf("unexpected memory: %+v", memory)
	}
	projects, err := ListProjects(c)
	assertOK(t, err)
	if len(projects) != 1 || projects[0].Name != "Hermdroid" {
		t.Fatalf("unexpected projects: %+v", projects)
	}
	tasks, err := ListKanbanTasks(c)
	assertOK(t, err)
	if len(tasks) != 1 || tasks[0].Status != "running" {
		t.Fatalf("unexpected tasks: %+v", tasks)
	}
	inbox, err := ListInboxItems(c)
	assertOK(t, err)
	if len(inbox) != 1 || inbox[0].Read {
		t.Fatalf("unexpected inbox: %+v", inbox)
	}
}

// TestLenientDecoding ensures malformed-but-recognizable fields don't kill parsing.
func TestLenientDecoding(t *testing.T) {
	payload := `{
		"id": "s1",
		"title": "Test",
		"input_tokens": "42",
		"pinned": "true",
		"estimated_cost": "0.00123"
	}`
	var s models.Session
	if err := json.Unmarshal([]byte(payload), &s); err != nil {
		t.Fatalf("lenient decode failed: %v", err)
	}
	if s.InputTokens != 42 {
		t.Errorf("input_tokens = %d, want 42", s.InputTokens)
	}
	if !s.Pinned {
		t.Error("pinned should be true")
	}
	if s.EstimatedCost < 0.001 || s.EstimatedCost > 0.002 {
		t.Errorf("estimated_cost = %v, want ~0.00123", s.EstimatedCost)
	}
}

// TestStreamMultipart ensures UploadFile builds a multipart body.
func TestUploadFileMultipart(t *testing.T) {
	var gotContentType string
	var body []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"filename":"up.txt"}`))
	}))
	defer ts.Close()
	c := client.NewAPIClientWithHTTP(ts.URL, ts.Client())
	c.SetBaseURL(ts.URL)
	_, err := UploadFile(c, &UploadFileRequest{SessionID: "s1", Filename: "up.txt", Content: []byte("data")})
	assertOK(t, err)
	if !strings.Contains(gotContentType, "multipart/form-data") {
		t.Errorf("content-type = %q, want multipart", gotContentType)
	}
	if !strings.Contains(string(body), "data") {
		t.Errorf("body missing uploaded content: %s", body)
	}
}
