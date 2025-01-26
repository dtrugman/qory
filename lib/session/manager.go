package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/dtrugman/qory/lib/message"
)

const (
	ValidIDPattern string = `^[a-zA-Z0-9_-]+$`
	ValidIDHint    string = "IDs may include letters [a-zA-Z], numbers [0-9], scores and underscores"
)

const (
	SessionsDirName = "sessions"

	sessionsFilePerm = 0600

	sessionPreviewChars = 64
)

var (
	ErrInvalidID = errors.New("invalid session id")
	ErrNotFound  = errors.New("unknown session id")
)

type Session struct {
	Messages []message.Message `json:"messages"`
}

func NewSession() Session {
	messages := make([]message.Message, 0)
	return Session{
		Messages: messages,
	}
}

func (s *Session) AddMessage(m message.Message) {
	s.Messages = append(s.Messages, m)
}

type SessionPreview struct {
	UpdatedAt time.Time
	Snippet   string
}

type fileInfo struct {
	name    string
	modTime time.Time
}

type manager struct {
	dir string
}

type Manager interface {
	Load(id string) (Session, error)
	Store(id string, session Session) error
	Enum(limit int) (map[string]SessionPreview, error)
}

func NewManager(dir string) (Manager, error) {
	return &manager{
		dir: dir,
	}, nil
}

func (m *manager) validID(id string) bool {
	pattern := "^[a-zA-Z0-9_-]+$"
	match, _ := regexp.MatchString(pattern, id)
	return match
}

func (m *manager) loadSessionSnippet(filename string) (string, error) {
	session, err := m.Load(filename)
	if err != nil {
		return "", err
	}

	var lastUserFound bool = false
	var lastUserContent string
	for i := len(session.Messages) - 1; i >= 0; i-- {
		if session.Messages[i].Role == message.RoleUser {
			lastUserFound = true
			lastUserContent = session.Messages[i].Content
			break
		}
	}

	if !lastUserFound {
		return "", fmt.Errorf("no user content")
	}

	content := lastUserContent
	if len(content) > sessionPreviewChars {
		content = content[:sessionPreviewChars]
		content += "..."
	}

	return content, nil
}

func (m *manager) Enum(limit int) (map[string]SessionPreview, error) {
	files, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %v", err)
	}

	fileInfos := make([]fileInfo, 0, len(files))
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			return nil, err
		}

		fileInfos = append(fileInfos, fileInfo{
			name:    f.Name(),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.After(fileInfos[j].modTime)
	})

	if limit > 0 && limit < len(fileInfos) {
		fileInfos = fileInfos[:limit]
	}

	result := make(map[string]SessionPreview)
	for _, info := range fileInfos {
		snippet, err := m.loadSessionSnippet(info.name)
		if err != nil {
			return nil, err
		}

		result[info.name] = SessionPreview{
			UpdatedAt: info.modTime,
			Snippet:   snippet,
		}
	}

	return result, nil
}

func (m *manager) Load(id string) (Session, error) {
	if !m.validID(id) {
		return Session{}, ErrInvalidID
	}

	var session Session

	path := filepath.Join(m.dir, id)
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, ErrNotFound
		} else {
			return Session{}, err
		}
	}

	err = json.Unmarshal(bytes, &session)
	return session, err
}

func (m *manager) Store(id string, session Session) error {
	if !m.validID(id) {
		return ErrInvalidID
	}

	b, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("encode: %v", err)
	}

	path := filepath.Join(m.dir, id)
	err = os.WriteFile(path, b, sessionsFilePerm)
	if err != nil {
		return fmt.Errorf("write file: %v", err)
	}

	return nil
}
