package biz

import (
	"errors"
	"fmt"

	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/message"
	"github.com/dtrugman/qory/lib/session"
	"github.com/google/uuid"
)

// Config is the interface for reading and writing persistent configuration.
type Config interface {
	GetConfigSubdir(name string) (string, error)

	Editor() (string, config.Origin, error)
	SetEditor(string) error
	UnsetEditor() error

	HistorySize() (int, config.Origin, error)
	SetHistorySize(string) error
	UnsetHistorySize() error

	Mode() (string, config.Origin, error)
	SetMode(string) error
	UnsetMode() error

	APIKey() (string, config.Origin, error)
	SetAPIKey(string) error
	UnsetAPIKey() error

	BaseURL() (string, config.Origin, error)
	SetBaseURL(string) error
	UnsetBaseURL() error

	Model() (string, config.Origin, error)
	SetModel(string) error
	UnsetModel() error

	Prompt() (string, config.Origin, error)
	SetPrompt(string) error
	UnsetPrompt() error
}

// Client is the interface for querying the language model.
type Client interface {
	AvailableModels() ([]string, error)
	Query(model string, messages []message.Message) (string, error)
}

// SessionManager is the interface for persisting chat sessions.
type SessionManager interface {
	Load(id string) (session.Session, error)
	Store(id string, s session.Session) error
	Delete(id string) error
	Enum(limit int) ([]session.SessionPreview, error)
	Last() (string, error)
	Cleanup(limit int) error
}

// Qory is the application object. All business logic lives here; Cobra
// command handlers are thin shims that delegate to these methods.
type Qory struct {
	conf   Config
	client Client
	sm     SessionManager
}

func NewQory(conf Config, client Client, sm SessionManager) *Qory {
	return &Qory{conf: conf, client: client, sm: sm}
}

// GetConfig returns the configuration object for direct access by callers.
func (q *Qory) GetConfig() Config {
	return q.conf
}

// runQueryInner is the shared query execution path. It appends the user prompt
// to sess, queries the model, and persists the updated session under sessionID.
func (q *Qory) runQueryInner(sessionID string, sess session.Session, inputs []string) error {
	modelName, _, err := q.conf.Model()
	if err != nil {
		return fmt.Errorf("get model failed: %w", err)
	}
	if modelName == "" {
		return fmt.Errorf("model is not set")
	}

	if len(sess.Messages) == 0 {
		systemPrompt, _, err := q.conf.Prompt()
		if err != nil {
			return fmt.Errorf("get system prompt failed: %w", err)
		}
		if systemPrompt != "" {
			sess.AddMessage(message.NewSystemMessage(systemPrompt))
		}
	}

	userPrompt := buildUserPrompt(inputs)
	sess.AddMessage(message.NewUserMessage(userPrompt))

	response, err := q.client.Query(modelName, sess.Messages)
	if err != nil {
		return err
	}

	sess.AddMessage(message.NewAssistantMessage(response))

	var errs []error
	if err = q.sm.Store(sessionID, sess); err != nil {
		errs = append(errs, fmt.Errorf("store session: %w", err))
	}
	historySize, _, err := q.conf.HistorySize()
	if err != nil {
		errs = append(errs, fmt.Errorf("get history size: %w", err))
	} else if err = q.sm.Cleanup(historySize); err != nil {
		errs = append(errs, fmt.Errorf("cleanup sessions: %w", err))
	}
	return errors.Join(errs...)
}

// QueryNew starts a fresh session with a new UUID. History is never loaded.
func (q *Qory) QueryNew(inputs []string) error {
	id := uuid.NewString()
	session := session.NewSession()
	return q.runQueryInner(id, session, inputs)
}

// QuerySession loads the session with the given ID (creating it if absent) and
// appends the new query to the existing conversation history.
func (q *Qory) QuerySession(id string, inputs []string) error {
	sess, err := q.sm.Load(id)
	if err != nil {
		return err
	}
	return q.runQueryInner(id, sess, inputs)
}

// QueryLast resolves the most recently modified session and continues it.
func (q *Qory) QueryLast(inputs []string) error {
	id, err := q.sm.Last()
	if err != nil {
		return err
	}
	return q.QuerySession(id, inputs)
}

// HistoryAll returns session previews. An optional limit caps the number
// returned; omit or pass 0 to return all sessions.
func (q *Qory) HistoryAll(limit int) ([]session.SessionPreview, error) {
	return q.sm.Enum(limit)
}

// HistorySession returns the full session for the given session ID.
func (q *Qory) HistorySession(sessionID string) (session.Session, error) {
	return q.sm.Load(sessionID)
}

// HistoryDelete deletes the session with the given ID.
func (q *Qory) HistoryDelete(sessionID string) error {
	return q.sm.Delete(sessionID)
}

// AvailableModels returns the list of models available from the client.
func (q *Qory) AvailableModels() ([]string, error) {
	return q.client.AvailableModels()
}

// QueryDefault runs a query using the configured default mode (new or last).
// If no mode is configured, it starts a new session.
func (q *Qory) QueryDefault(inputs []string) error {
	mode, _, err := q.conf.Mode()
	if err != nil {
		return fmt.Errorf("get mode failed: %w", err)
	}
	if mode == config.ModeLast {
		return q.QueryLast(inputs)
	}
	return q.QueryNew(inputs)
}
