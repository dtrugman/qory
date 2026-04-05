package biz

import (
	"errors"
	"testing"
	"time"

	"github.com/dtrugman/qory/lib/message"
	"github.com/dtrugman/qory/lib/session"
	"github.com/dtrugman/qory/lib/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---- mock config ----

type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetConfigSubdir(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

func (m *MockConfig) Get(key string) (*string, error) {
	args := m.Called(key)
	val, _ := args.Get(0).(*string)
	return val, args.Error(1)
}

func (m *MockConfig) Set(key string, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockConfig) Unset(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// ---- mock client ----

type MockClient struct {
	mock.Mock
}

func (m *MockClient) AvailableModels() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockClient) Query(model string, msgs []message.Message) (string, error) {
	args := m.Called(model, msgs)
	return args.String(0), args.Error(1)
}

// ---- mock session manager ----

type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) Load(id string) (session.Session, error) {
	args := m.Called(id)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *MockSessionManager) Store(id string, s session.Session) error {
	args := m.Called(id, s)
	return args.Error(0)
}

func (m *MockSessionManager) Enum(limit int) ([]session.SessionPreview, error) {
	args := m.Called(limit)
	return args.Get(0).([]session.SessionPreview), args.Error(1)
}

func (m *MockSessionManager) Last() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockSessionManager) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSessionManager) Cleanup(limit int) error {
	args := m.Called(limit)
	return args.Error(0)
}

// ---- QueryNew tests ----

func Test_QueryNew_DoesNotLoadHistory(t *testing.T) {
	userText := "hello"
	assistantText := "response"

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	conf.On("Get", "prompt").Return((*string)(nil), nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", mock.AnythingOfType("string"), expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QueryNew([]string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QueryNew_UsesUniqueSessionIDs(t *testing.T) {
	firstUserText := "first"
	secondUserText := "second"
	assistantText := "response"

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	conf.On("Get", "prompt").Return((*string)(nil), nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(firstUserText),
	}).Return(assistantText, nil).Once()
	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(secondUserText),
	}).Return(assistantText, nil).Once()
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	firstExpected := session.NewSession()
	firstExpected.AddMessage(message.NewUserMessage(firstUserText))
	firstExpected.AddMessage(message.NewAssistantMessage(assistantText))

	secondExpected := session.NewSession()
	secondExpected.AddMessage(message.NewUserMessage(secondUserText))
	secondExpected.AddMessage(message.NewAssistantMessage(assistantText))

	var storedIDs []string
	captureID := func(args mock.Arguments) {
		storedIDs = append(storedIDs, args.String(0))
	}

	sm.On("Store", mock.AnythingOfType("string"), firstExpected).
		Run(captureID).Return(nil).Once()
	sm.On("Store", mock.AnythingOfType("string"), secondExpected).
		Run(captureID).Return(nil).Once()

	q := NewQory(conf, client, sm)
	err1 := q.QueryNew([]string{firstUserText})
	require.NoError(t, err1)
	err2 := q.QueryNew([]string{secondUserText})
	require.NoError(t, err2)

	require.Len(t, storedIDs, 2)
	assert.NotEqual(t, storedIDs[0], storedIDs[1])

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QueryNew_AddsSystemPrompt(t *testing.T) {
	systemText := "Be concise."
	userText := "hello"
	assistantText := "response"

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	conf.On("Get", "prompt").Return(util.Ptr(systemText), nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewSystemMessage(systemText),
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewSystemMessage(systemText))
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", mock.AnythingOfType("string"), expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QueryNew([]string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QueryNew_ReturnsErrorOnQueryFailure(t *testing.T) {
	queryErr := errors.New("rate limit exceeded")

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	conf.On("Get", "prompt").Return((*string)(nil), nil)

	client.On("Query", "gpt-4o", mock.Anything).Return("", queryErr)

	q := NewQory(conf, client, sm)
	err := q.QueryNew([]string{"hello"})
	require.ErrorIs(t, err, queryErr)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

// ---- QuerySession tests ----

func Test_QuerySession_LoadsExistingHistory(t *testing.T) {
	prevUserText := "previous question"
	prevAssistantText := "previous answer"
	userText := "follow up"
	assistantText := "new response"

	existing := session.NewSession()
	existing.AddMessage(message.NewUserMessage(prevUserText))
	existing.AddMessage(message.NewAssistantMessage(prevAssistantText))

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	sm.On("Load", "my-session").Return(existing, nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(prevUserText),
		message.NewAssistantMessage(prevAssistantText),
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewUserMessage(prevUserText))
	expectedSession.AddMessage(message.NewAssistantMessage(prevAssistantText))
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", "my-session", expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QuerySession("my-session", []string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QuerySession_FailsWhenNotFound(t *testing.T) {
	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	sm.On("Load", "unknown-session").Return(session.Session{}, session.ErrNotFound)

	q := NewQory(conf, client, sm)
	err := q.QuerySession("unknown-session", []string{"hello"})

	require.ErrorIs(t, err, session.ErrNotFound)
	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QuerySession_SkipsSystemPrompt(t *testing.T) {
	systemText := "Be concise."
	prevUserText := "previous"
	prevAssistantText := "previous answer"
	userText := "follow up"
	assistantText := "response"

	existing := session.NewSession()
	existing.AddMessage(message.NewSystemMessage(systemText))
	existing.AddMessage(message.NewUserMessage(prevUserText))
	existing.AddMessage(message.NewAssistantMessage(prevAssistantText))

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	sm.On("Load", "my-session").Return(existing, nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewSystemMessage(systemText),
		message.NewUserMessage(prevUserText),
		message.NewAssistantMessage(prevAssistantText),
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewSystemMessage(systemText))
	expectedSession.AddMessage(message.NewUserMessage(prevUserText))
	expectedSession.AddMessage(message.NewAssistantMessage(prevAssistantText))
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", "my-session", expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QuerySession("my-session", []string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

// ---- QueryLast tests ----

func Test_QueryLast_ResolvesLastSessionID(t *testing.T) {
	prevUserText := "previous"
	prevAssistantText := "answer"
	userText := "follow up"
	assistantText := "response"

	existing := session.NewSession()
	existing.AddMessage(message.NewUserMessage(prevUserText))
	existing.AddMessage(message.NewAssistantMessage(prevAssistantText))

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	sm.On("Last").Return("last-session", nil)
	sm.On("Load", "last-session").Return(existing, nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(prevUserText),
		message.NewAssistantMessage(prevAssistantText),
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewUserMessage(prevUserText))
	expectedSession.AddMessage(message.NewAssistantMessage(prevAssistantText))
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", "last-session", expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QueryLast([]string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QueryLast_SkipsSystemPrompt(t *testing.T) {
	systemText := "Be concise."
	prevUserText := "previous"
	prevAssistantText := "previous answer"
	userText := "follow up"
	assistantText := "response"

	existing := session.NewSession()
	existing.AddMessage(message.NewSystemMessage(systemText))
	existing.AddMessage(message.NewUserMessage(prevUserText))
	existing.AddMessage(message.NewAssistantMessage(prevAssistantText))

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	sm.On("Last").Return("last-session", nil)
	sm.On("Load", "last-session").Return(existing, nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewSystemMessage(systemText),
		message.NewUserMessage(prevUserText),
		message.NewAssistantMessage(prevAssistantText),
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewSystemMessage(systemText))
	expectedSession.AddMessage(message.NewUserMessage(prevUserText))
	expectedSession.AddMessage(message.NewAssistantMessage(prevAssistantText))
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", "last-session", expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QueryLast([]string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

// ---- QueryDefault tests ----

func Test_QueryDefault_NoConfig(t *testing.T) {
	userText := "hello"
	assistantText := "response"

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "mode").Return((*string)(nil), nil)
	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	conf.On("Get", "prompt").Return((*string)(nil), nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", mock.AnythingOfType("string"), expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QueryDefault([]string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QueryDefault_ModeNew(t *testing.T) {
	userText := "hello"
	assistantText := "response"

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "mode").Return(util.Ptr("new"), nil)
	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	conf.On("Get", "prompt").Return((*string)(nil), nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", mock.AnythingOfType("string"), expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QueryDefault([]string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_QueryDefault_ModeLast(t *testing.T) {
	prevUserText := "previous"
	prevAssistantText := "answer"
	userText := "follow up"
	assistantText := "response"

	existing := session.NewSession()
	existing.AddMessage(message.NewUserMessage(prevUserText))
	existing.AddMessage(message.NewAssistantMessage(prevAssistantText))

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	conf.On("Get", "mode").Return(util.Ptr("last"), nil)
	conf.On("Get", "model").Return(util.Ptr("gpt-4o"), nil)
	sm.On("Last").Return("last-session", nil)
	sm.On("Load", "last-session").Return(existing, nil)

	client.On("Query", "gpt-4o", []message.Message{
		message.NewUserMessage(prevUserText),
		message.NewAssistantMessage(prevAssistantText),
		message.NewUserMessage(userText),
	}).Return(assistantText, nil)

	expectedSession := session.NewSession()
	expectedSession.AddMessage(message.NewUserMessage(prevUserText))
	expectedSession.AddMessage(message.NewAssistantMessage(prevAssistantText))
	expectedSession.AddMessage(message.NewUserMessage(userText))
	expectedSession.AddMessage(message.NewAssistantMessage(assistantText))
	sm.On("Store", "last-session", expectedSession).Return(nil)
	sm.On("Cleanup", sessionUnnamedLimit).Return(nil)

	q := NewQory(conf, client, sm)
	err := q.QueryDefault([]string{userText})
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

// ---- History tests ----

func Test_History_AllCallsEnum(t *testing.T) {
	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	previews := []session.SessionPreview{
		{Name: "test-session", UpdatedAt: time.Now(), Snippet: "test snippet"},
	}
	sm.On("Enum", historyLength).Return(previews, nil)

	q := NewQory(conf, client, sm)
	result, err := q.HistoryAll()
	require.NoError(t, err)
	assert.Equal(t, previews, result)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_History_SessionLoadsCorrectSession(t *testing.T) {
	existing := session.NewSession()
	existing.AddMessage(message.NewUserMessage("test\n"))

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	sm.On("Load", "my-session").Return(existing, nil)

	q := NewQory(conf, client, sm)
	result, err := q.HistorySession("my-session")
	require.NoError(t, err)
	assert.Equal(t, existing, result)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_History_DeleteCallsDelete(t *testing.T) {
	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	sm.On("Delete", "my-session").Return(nil)

	q := NewQory(conf, client, sm)
	err := q.HistoryDelete("my-session")
	require.NoError(t, err)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_History_DeleteReturnsError(t *testing.T) {
	deleteErr := errors.New("not found")

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	sm.On("Delete", "missing-session").Return(deleteErr)

	q := NewQory(conf, client, sm)
	err := q.HistoryDelete("missing-session")
	require.ErrorIs(t, err, deleteErr)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_History_SessionNotFoundReturnsError(t *testing.T) {
	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	sm.On("Load", "nonexistent").Return(session.Session{}, session.ErrNotFound)

	q := NewQory(conf, client, sm)
	_, err := q.HistorySession("nonexistent")
	require.ErrorIs(t, err, session.ErrNotFound)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

// ---- Config tests ----

func Test_Config_GetRoutesCorrectKey(t *testing.T) {
	cases := []struct {
		name        string
		fn          func(*Qory) (*string, error)
		expectedKey string
	}{
		{"api-key", (*Qory).ConfigGetAPIKey, "api_key"},
		{"base-url", (*Qory).ConfigGetBaseURL, "base_url"},
		{"model", (*Qory).ConfigGetModel, "model"},
		{"prompt", (*Qory).ConfigGetPrompt, "prompt"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expected := util.Ptr("test-value")
			conf := &MockConfig{}
			conf.On("Get", tc.expectedKey).Return(expected, nil)

			q := NewQory(conf, &MockClient{}, &MockSessionManager{})
			result, err := tc.fn(q)
			require.NoError(t, err)
			assert.Equal(t, expected, result)

			conf.AssertExpectations(t)
		})
	}
}

func Test_Config_UnsetRoutesCorrectKey(t *testing.T) {
	cases := []struct {
		name        string
		fn          func(*Qory) error
		expectedKey string
	}{
		{"api-key", (*Qory).ConfigUnsetAPIKey, "api_key"},
		{"base-url", (*Qory).ConfigUnsetBaseURL, "base_url"},
		{"model", (*Qory).ConfigUnsetModel, "model"},
		{"prompt", (*Qory).ConfigUnsetPrompt, "prompt"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conf := &MockConfig{}
			conf.On("Unset", tc.expectedKey).Return(nil)

			q := NewQory(conf, &MockClient{}, &MockSessionManager{})
			err := tc.fn(q)
			require.NoError(t, err)

			conf.AssertExpectations(t)
		})
	}
}

func Test_Config_SetRoutesCorrectKey(t *testing.T) {
	cases := []struct {
		name        string
		fn          func(*Qory, string) error
		expectedKey string
		value       string
	}{
		{"api-key", (*Qory).ConfigSetAPIKey, "api_key", "sk-test-key"},
		{"base-url", (*Qory).ConfigSetBaseURL, "base_url", "https://api.example.com/"},
		{"model", (*Qory).ConfigSetModel, "model", "gpt-4o"},
		{"prompt", (*Qory).ConfigSetPrompt, "prompt", "Be concise."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conf := &MockConfig{}
			conf.On("Set", tc.expectedKey, tc.value).Return(nil)

			q := NewQory(conf, &MockClient{}, &MockSessionManager{})
			err := tc.fn(q, tc.value)
			require.NoError(t, err)

			conf.AssertExpectations(t)
		})
	}
}

func Test_Config_SetBaseURLNormalizesTrailingSlash(t *testing.T) {
	conf := &MockConfig{}
	conf.On("Set", "base_url", "https://api.example.com/").Return(nil)

	q := NewQory(conf, &MockClient{}, &MockSessionManager{})
	err := q.ConfigSetBaseURL("https://api.example.com")
	require.NoError(t, err)

	conf.AssertExpectations(t)
}

func Test_Config_GetRoutesCorrectKey_Mode(t *testing.T) {
	expected := util.Ptr("new")
	conf := &MockConfig{}
	conf.On("Get", "mode").Return(expected, nil)

	q := NewQory(conf, &MockClient{}, &MockSessionManager{})
	result, err := q.ConfigGetMode()
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	conf.AssertExpectations(t)
}

func Test_Config_UnsetRoutesCorrectKey_Mode(t *testing.T) {
	conf := &MockConfig{}
	conf.On("Unset", "mode").Return(nil)

	q := NewQory(conf, &MockClient{}, &MockSessionManager{})
	err := q.ConfigUnsetMode()
	require.NoError(t, err)

	conf.AssertExpectations(t)
}

func Test_Config_SetMode_AcceptsValidValues(t *testing.T) {
	for _, value := range []string{"new", "last"} {
		t.Run(value, func(t *testing.T) {
			conf := &MockConfig{}
			conf.On("Set", "mode", value).Return(nil)

			q := NewQory(conf, &MockClient{}, &MockSessionManager{})
			err := q.ConfigSetMode(value)
			require.NoError(t, err)

			conf.AssertExpectations(t)
		})
	}
}

func Test_Config_SetMode_RejectsInvalidValue(t *testing.T) {
	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	q := NewQory(conf, client, sm)
	err := q.ConfigSetMode("invalid")
	require.Error(t, err)

	conf.AssertExpectations(t)
	client.AssertExpectations(t)
	sm.AssertExpectations(t)
}

func Test_Config_GetRoutesCorrectKey_Editor(t *testing.T) {
	expected := util.Ptr("nvim")
	conf := &MockConfig{}
	conf.On("Get", "editor").Return(expected, nil)

	q := NewQory(conf, &MockClient{}, &MockSessionManager{})
	result, err := q.ConfigGetEditor()
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	conf.AssertExpectations(t)
}

func Test_Config_UnsetRoutesCorrectKey_Editor(t *testing.T) {
	conf := &MockConfig{}
	conf.On("Unset", "editor").Return(nil)

	q := NewQory(conf, &MockClient{}, &MockSessionManager{})
	err := q.ConfigUnsetEditor()
	require.NoError(t, err)

	conf.AssertExpectations(t)
}

func Test_Config_SetRoutesCorrectKey_Editor(t *testing.T) {
	conf := &MockConfig{}
	conf.On("Set", "editor", "nvim").Return(nil)

	q := NewQory(conf, &MockClient{}, &MockSessionManager{})
	err := q.ConfigSetEditor("nvim")
	require.NoError(t, err)

	conf.AssertExpectations(t)
}

// ---- QueryLast error tests ----

func Test_QueryLast_FailsWhenLastErrors(t *testing.T) {
	lastErr := errors.New("no sessions")

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	sm.On("Last").Return("", lastErr)

	q := NewQory(conf, client, sm)
	err := q.QueryLast([]string{"hello"})
	require.ErrorIs(t, err, lastErr)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

// ---- AvailableModels tests ----

func Test_AvailableModels_ReturnsClientModels(t *testing.T) {
	models := []string{"gpt-4o", "gpt-4o-mini"}

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	client.On("AvailableModels").Return(models, nil)

	q := NewQory(conf, client, sm)
	result, err := q.AvailableModels()
	require.NoError(t, err)
	assert.Equal(t, models, result)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}

func Test_AvailableModels_ReturnsErrorOnFailure(t *testing.T) {
	modelsErr := errors.New("unauthorized")

	conf := &MockConfig{}
	client := &MockClient{}
	sm := &MockSessionManager{}

	client.On("AvailableModels").Return([]string{}, modelsErr)

	q := NewQory(conf, client, sm)
	_, err := q.AvailableModels()
	require.ErrorIs(t, err, modelsErr)

	sm.AssertExpectations(t)
	conf.AssertExpectations(t)
	client.AssertExpectations(t)
}
