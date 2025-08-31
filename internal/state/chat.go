package state

// Exchange represents a single chat exchange
type Exchange struct {
	Prompt   string
	Response string
}

// ChatState manages the chat conversation state
type ChatState struct {
	history       []Exchange
	thinking      bool
	currentPrompt string
	err           error
}

// NewChatState creates a new chat state manager
func NewChatState() *ChatState {
	return &ChatState{
		history: []Exchange{},
	}
}

// AddExchange adds a new exchange to the history
func (cs *ChatState) AddExchange(prompt, response string) {
	cs.history = append(cs.history, Exchange{
		Prompt:   prompt,
		Response: response,
	})
}

// GetHistory returns the chat history
func (cs *ChatState) GetHistory() []Exchange {
	return cs.history
}

// ClearHistory clears the chat history
func (cs *ChatState) ClearHistory() {
	cs.history = []Exchange{}
}

// SetThinking sets the thinking state
func (cs *ChatState) SetThinking(thinking bool) {
	cs.thinking = thinking
}

// IsThinking returns whether the chat is currently thinking
func (cs *ChatState) IsThinking() bool {
	return cs.thinking
}

// SetCurrentPrompt sets the current prompt being processed
func (cs *ChatState) SetCurrentPrompt(prompt string) {
	cs.currentPrompt = prompt
}

// GetCurrentPrompt returns the current prompt
func (cs *ChatState) GetCurrentPrompt() string {
	return cs.currentPrompt
}

// ClearCurrentPrompt clears the current prompt
func (cs *ChatState) ClearCurrentPrompt() {
	cs.currentPrompt = ""
}

// SetError sets the current error
func (cs *ChatState) SetError(err error) {
	cs.err = err
}

// GetError returns the current error
func (cs *ChatState) GetError() error {
	return cs.err
}

// ClearError clears the current error
func (cs *ChatState) ClearError() {
	cs.err = nil
}

// GetMessageCount returns the number of messages in history
func (cs *ChatState) GetMessageCount() int {
	return len(cs.history)
}
