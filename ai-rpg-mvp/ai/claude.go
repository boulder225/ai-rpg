package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider implements the AIProvider interface using Claude API
type ClaudeProvider struct {
	client      anthropic.Client
	model       string
	maxTokens   int64
	temperature float64
	timeout     time.Duration
}

// NewClaudeProvider creates a new Claude AI provider
func NewClaudeProvider(config AIConfig) (*ClaudeProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	client := anthropic.NewClient(option.WithAPIKey(config.APIKey))

	// Use model string directly
	model := config.Model
	if model == "" {
		model = "claude-3-sonnet-20240229"
	}

	maxTokens := int64(config.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 1000
	}

	temperature := config.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ClaudeProvider{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		timeout:     timeout,
	}, nil
}

// GenerateGMResponse generates a Game Master response using Claude
func (c *ClaudeProvider) GenerateGMResponse(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Enhance the prompt with GM-specific instructions
	systemPrompt := `You are an expert AI Game Master running a fantasy RPG session. Your role:

PERSONALITY: Helpful yet challenging guide who creates immersive experiences
TONE: Descriptive, engaging, appropriate to fantasy setting  
GOALS: Player agency, narrative flow, consistent world-building

RESPONSE GUIDELINES:
- Always respond in character as the GM
- Maintain world consistency across interactions
- React contextually to player actions and equipment  
- Balance guidance with player discovery
- Generate consequences for player choices
- Keep responses engaging and immersive (2-4 sentences)
- End with a clear situation that allows player response

Current game situation requires your response as Game Master.`

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens,
		System:    []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Temperature: anthropic.Float(c.temperature),
	})

	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return message.Content[0].Text, nil
}

// GenerateNPCDialogue generates NPC dialogue using Claude
func (c *ClaudeProvider) GenerateNPCDialogue(npcName, personality, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	systemPrompt := fmt.Sprintf(`You are %s, an NPC in a fantasy RPG world.

PERSONALITY TRAITS: %s

DIALOGUE GUIDELINES:
- Stay in character as %s at all times
- Speak naturally and authentically for this character
- Reference your personality and background
- Respond appropriately to the player's actions and reputation
- Keep dialogue concise but meaningful (1-3 sentences)
- Include personality quirks or speech patterns
- Consider your relationship with the player

Respond as %s would naturally speak in this situation.`,
		npcName, personality, npcName, npcName)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens / 2, // Shorter responses for NPCs
		System:    []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Temperature: anthropic.Float(c.temperature + 0.1), // Slightly more creative for NPCs
	})

	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return message.Content[0].Text, nil
}

// GenerateSceneDescription generates scene descriptions using Claude
func (c *ClaudeProvider) GenerateSceneDescription(location, contextInfo, mood string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	systemPrompt := `You are a skilled fantasy writer creating immersive scene descriptions for an RPG.

DESCRIPTION GUIDELINES:
- Create vivid, atmospheric descriptions that set the mood
- Include sensory details (sight, sound, smell, feel)
- Match the tone and mood of the situation
- Keep descriptions concise but evocative (2-3 sentences)
- Focus on elements that enhance gameplay and immersion
- Include details that suggest possible interactions or discoveries
- Maintain consistency with fantasy RPG conventions

Create an engaging scene description based on the provided context.`

	scenePrompt := fmt.Sprintf(`Location: %s
Context: %s
Mood/Atmosphere: %s

Describe this scene:`, location, contextInfo, mood)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens / 2,
		System:    []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(scenePrompt)),
		},
		Temperature: anthropic.Float(c.temperature + 0.2), // More creative for descriptions
	})

	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return message.Content[0].Text, nil
}

// GetProviderName returns the provider name
func (c *ClaudeProvider) GetProviderName() string {
	return "claude"
}

// ValidateClaudeConfig validates Claude-specific configuration
func ValidateClaudeConfig(config AIConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("Claude API key is required")
	}
	return nil
}
