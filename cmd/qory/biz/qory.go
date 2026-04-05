package biz

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/message"
	"github.com/dtrugman/qory/lib/session"
	"github.com/google/uuid"
)

const (
	appName = "Qory"

	historyLength      = 10
	defaultHistorySize = 50

	defaultEditor = "vi"
)

var version = "dev"

// Config is the interface for reading and writing persistent configuration.
type Config interface {
	GetConfigSubdir(name string) (string, error)
	Get(key string) (*string, error)
	Set(key string, value string) error
	Unset(key string) error
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

// runQueryInner is the shared query execution path. It appends the user prompt
// to sess, queries the model, and persists the updated session under sessionID.
func (q *Qory) runQueryInner(sessionID string, sess session.Session, inputs []string) error {
	modelName, err := q.configGet(config.Model)
	if err != nil {
		return fmt.Errorf("get model failed: %w", err)
	}
	if modelName == nil {
		return fmt.Errorf("model is not set")
	}

	if len(sess.Messages) == 0 {
		systemPrompt, err := q.configGet(config.Prompt)
		if err != nil {
			return fmt.Errorf("get system prompt failed: %w", err)
		}
		if systemPrompt != nil {
			sess.AddMessage(message.NewSystemMessage(*systemPrompt))
		}
	}

	userPrompt := buildUserPrompt(inputs)
	sess.AddMessage(message.NewUserMessage(userPrompt))

	response, err := q.client.Query(*modelName, sess.Messages)
	if err != nil {
		return err
	}

	sess.AddMessage(message.NewAssistantMessage(response))

	var errs []error
	if err = q.sm.Store(sessionID, sess); err != nil {
		errs = append(errs, fmt.Errorf("store session: %w", err))
	}
	historySize, err := q.getHistorySize()
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

// HistoryAll returns a preview for each of the most recent sessions.
func (q *Qory) HistoryAll() ([]session.SessionPreview, error) {
	return q.sm.Enum(historyLength)
}

// HistorySession returns the full session for the given session ID.
func (q *Qory) HistorySession(sessionID string) (session.Session, error) {
	return q.sm.Load(sessionID)
}

// HistoryDelete deletes the session with the given ID.
func (q *Qory) HistoryDelete(sessionID string) error {
	return q.sm.Delete(sessionID)
}

// Version returns the application version string.
func (q *Qory) Version() string {
	return fmt.Sprintf("%s version %s", appName, version)
}

// AvailableModels returns the list of models available from the client.
func (q *Qory) AvailableModels() ([]string, error) {
	return q.client.AvailableModels()
}

// configGet returns the stored value for key, or nil if unset.
func (q *Qory) configGet(key string) (*string, error) {
	return q.conf.Get(key)
}

// configUnset removes the stored value for key.
func (q *Qory) configUnset(key string) error {
	return q.conf.Unset(key)
}

// configSet stores value for key.
func (q *Qory) configSet(key string, value string) error {
	return q.conf.Set(key, value)
}

func (q *Qory) ConfigGetAPIKey() (*string, error) {
	return q.configGet(config.APIKey)
}

func (q *Qory) ConfigSetAPIKey(value string) error {
	return q.configSet(config.APIKey, value)
}

func (q *Qory) ConfigUnsetAPIKey() error {
	return q.configUnset(config.APIKey)
}

func (q *Qory) ConfigGetBaseURL() (*string, error) {
	return q.configGet(config.BaseURL)
}

func (q *Qory) ConfigSetBaseURL(value string) error {
	if !strings.HasSuffix(value, "/") {
		value = value + "/"
	}
	return q.configSet(config.BaseURL, value)
}

func (q *Qory) ConfigUnsetBaseURL() error {
	return q.configUnset(config.BaseURL)
}

func (q *Qory) ConfigGetModel() (*string, error) {
	return q.configGet(config.Model)
}

func (q *Qory) ConfigSetModel(value string) error {
	return q.configSet(config.Model, value)
}

func (q *Qory) ConfigUnsetModel() error {
	return q.configUnset(config.Model)
}

func (q *Qory) ConfigGetPrompt() (*string, error) {
	return q.configGet(config.Prompt)
}

func (q *Qory) ConfigSetPrompt(value string) error {
	return q.configSet(config.Prompt, value)
}

func (q *Qory) ConfigUnsetPrompt() error {
	return q.configUnset(config.Prompt)
}

func (q *Qory) ConfigGetMode() (*string, error) {
	return q.configGet(config.Mode)
}

func (q *Qory) ConfigSetMode(value string) error {
	switch value {
	case config.ModeNew, config.ModeLast:
		// Valid

	default:
		return fmt.Errorf("invalid mode %q", value)
	}

	return q.configSet(config.Mode, value)
}

func (q *Qory) ConfigUnsetMode() error {
	return q.configUnset(config.Mode)
}

func (q *Qory) ConfigGetEditor() (*string, error) {
	return q.configGet(config.Editor)
}

func (q *Qory) ConfigSetEditor(value string) error {
	return q.configSet(config.Editor, value)
}

func (q *Qory) ConfigUnsetEditor() error {
	return q.configUnset(config.Editor)
}

func (q *Qory) ConfigGetHistorySizeDefault() int {
	return defaultHistorySize
}

func (q *Qory) ConfigGetHistorySize() (*string, error) {
	return q.configGet(config.HistorySize)
}

func (q *Qory) ConfigSetHistorySize(value string) error {
	size, err := strconv.Atoi(value)
	if err != nil || size <= 0 {
		return fmt.Errorf("invalid history size %q: must be a positive integer", value)
	}

	return q.configSet(config.HistorySize, value)
}

func (q *Qory) ConfigUnsetHistorySize() error {
	return q.configUnset(config.HistorySize)
}

// getHistorySize returns the configured number of unnamed sessions to retain,
// falling back to sessionUnnamedLimit if no value is set.
func (q *Qory) getHistorySize() (int, error) {
	v, err := q.configGet(config.HistorySize)
	if err != nil {
		return 0, fmt.Errorf("get history size failed: %w", err)
	}

	if v == nil {
		return defaultHistorySize, nil
	}

	size, err := strconv.Atoi(*v)
	if err != nil {
		return 0, fmt.Errorf("invalid history size %q: %w", *v, err)
	}

	return size, nil
}

// GetEditor resolves the editor to use: config → $VISUAL → $EDITOR → defaultEditor.
func (q *Qory) GetEditor() (string, error) {
	v, err := q.configGet(config.Editor)
	if err != nil {
		return "", fmt.Errorf("get editor failed: %w", err)
	}
	if v != nil && *v != "" {
		return *v, nil
	}
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual, nil
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed, nil
	}
	return defaultEditor, nil
}

// QueryDefault runs a query using the configured default mode (new or last).
// If no mode is configured, it starts a new session.
func (q *Qory) QueryDefault(inputs []string) error {
	mode, err := q.configGet(config.Mode)
	if err != nil {
		return fmt.Errorf("get mode failed: %w", err)
	}
	if mode != nil && *mode == config.ModeLast {
		return q.QueryLast(inputs)
	}
	return q.QueryNew(inputs)
}
