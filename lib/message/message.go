package message

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

func NewRoleMessage(role Role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
	}
}

func NewUserMessage(content string) Message {
	return NewRoleMessage(RoleUser, content)
}

func NewSystemMessage(content string) Message {
	return NewRoleMessage(RoleSystem, content)
}

func NewAssistantMessage(content string) Message {
	return NewRoleMessage(RoleAssistant, content)
}
