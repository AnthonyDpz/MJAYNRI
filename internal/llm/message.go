package llm

// Role identifie l'auteur d'un message dans une conversation.
type Role string

const (
	// RoleUser désigne un message envoyé par l'utilisateur.
	RoleUser Role = "user"
	// RoleAssistant désigne une réponse de l'IA.
	RoleAssistant Role = "assistant"
	// RoleSystem désigne une instruction système (prompt système).
	RoleSystem Role = "system"
)

// Message représente un tour de conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Chunk est un fragment de réponse reçu en streaming.
type Chunk struct {
	// Content est le texte partiel reçu.
	Content string
	// Done indique que la réponse est terminée.
	Done bool
	// Err contient une éventuelle erreur de streaming.
	Err error
}
