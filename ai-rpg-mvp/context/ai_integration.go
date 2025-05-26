package context

import (
	"fmt"
	"strings"
	"time"
)

// GetContextSummary generates a summary for AI integration
func (cm *ContextManager) GetContextSummary(sessionID string) (*ContextSummary, error) {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return nil, err
	}

	summary := &ContextSummary{
		CurrentLocation:    ctx.Location.Current,
		PreviousLocation:   ctx.Location.Previous,
		PlayerHealth:       fmt.Sprintf("%d/%d", ctx.Character.Health.Current, ctx.Character.Health.Max),
		PlayerReputation:   ctx.Character.Reputation,
		RecentActions:      cm.getActionSummary(ctx.Actions, 5),
		ActiveNPCs:         cm.getRelevantNPCs(ctx),
		SessionDuration:    time.Since(ctx.StartTime).Minutes(),
		PlayerMood:         cm.determinePlayerMood(ctx),
		WorldState:         make(map[string]interface{}),
	}

	// Add world state information
	summary.WorldState["locations_visited"] = ctx.SessionStats.LocationsVisited
	summary.WorldState["total_actions"] = ctx.SessionStats.TotalActions
	summary.WorldState["combat_experienced"] = ctx.SessionStats.CombatActions > 0
	summary.WorldState["social_active"] = ctx.SessionStats.SocialActions > ctx.SessionStats.CombatActions

	return summary, nil
}

// GenerateAIPrompt creates a structured prompt for the AI GM
func (cm *ContextManager) GenerateAIPrompt(sessionID string) (string, error) {
	summary, err := cm.GetContextSummary(sessionID)
	if err != nil {
		return "", err
	}

	recentActions, err := cm.GetRecentActions(sessionID, 3)
	if err != nil {
		return "", err
	}

	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`GAME MASTER CONTEXT

CURRENT GAME STATE:
- Location: %s (previously: %s)
- Player Health: %s
- Player Reputation: %d (%s)
- Session Duration: %.1f minutes
- Player Mood: %s

RECENT PLAYER ACTIONS:
%s

ACTIVE NPCS IN AREA:
%s

PLAYER CHARACTER:
- Name: %s
- Equipment: %s
- Recent Focus: %s

WORLD CONTEXT:
%s

GM INSTRUCTIONS:
You are the AI Game Master for this fantasy RPG session. Based on the current context:
1. Respond as the omniscient narrator and world
2. Maintain consistency with previous interactions
3. React appropriately to the player's reputation and recent actions
4. Consider NPC relationships and dispositions
5. Provide immersive, contextual descriptions
6. Balance challenge with player agency

Current situation requires your response as Game Master.`,
		summary.CurrentLocation,
		cm.formatPreviousLocation(summary.PreviousLocation),
		summary.PlayerHealth,
		summary.PlayerReputation,
		cm.getReputationDescription(summary.PlayerReputation),
		summary.SessionDuration,
		summary.PlayerMood,
		cm.formatRecentActions(recentActions),
		cm.formatActiveNPCs(summary.ActiveNPCs),
		ctx.Character.Name,
		cm.formatEquipment(ctx.Character.Equipment),
		cm.determinePlayerFocus(ctx),
		cm.formatWorldContext(summary.WorldState),
	)

	return prompt, nil
}

// GenerateAIPromptData creates structured data for advanced AI integration
func (cm *ContextManager) GenerateAIPromptData(sessionID string) (*AIPromptData, error) {
	summary, err := cm.GetContextSummary(sessionID)
	if err != nil {
		return nil, err
	}

	recentEvents, err := cm.GetRecentActions(sessionID, 10)
	if err != nil {
		return nil, err
	}

	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return nil, err
	}

	promptData := &AIPromptData{
		SessionContext: summary,
		RecentEvents:   recentEvents,
		WorldKnowledge: make(map[string]interface{}),
		PlayerProfile:  make(map[string]interface{}),
		GMPersonality:  make(map[string]interface{}),
	}

	// Player profile
	promptData.PlayerProfile["name"] = ctx.Character.Name
	promptData.PlayerProfile["play_style"] = cm.determinePlayStyle(ctx)
	promptData.PlayerProfile["experience_level"] = cm.determineExperienceLevel(ctx)
	promptData.PlayerProfile["preferred_activities"] = cm.getPreferredActivities(ctx)

	// GM personality configuration
	promptData.GMPersonality["helpfulness"] = 0.7
	promptData.GMPersonality["challenge_level"] = 0.6
	promptData.GMPersonality["mystery_level"] = 0.6
	promptData.GMPersonality["immersion_focus"] = 0.9

	// World knowledge
	promptData.WorldKnowledge["known_locations"] = cm.getKnownLocations(ctx)
	promptData.WorldKnowledge["established_npcs"] = cm.getEstablishedNPCs(ctx)
	promptData.WorldKnowledge["ongoing_storylines"] = cm.getOngoingStorylines(ctx)

	return promptData, nil
}

// Helper functions for AI prompt generation

func (cm *ContextManager) getActionSummary(actions []ActionEvent, count int) []string {
	if len(actions) == 0 {
		return []string{"No recent actions"}
	}

	startIdx := len(actions) - count
	if startIdx < 0 {
		startIdx = 0
	}

	summaries := make([]string, 0, count)
	for i := startIdx; i < len(actions); i++ {
		action := actions[i]
		summary := fmt.Sprintf("%s: %s -> %s", 
			action.Type, 
			action.Command, 
			action.Outcome)
		summaries = append(summaries, summary)
	}

	return summaries
}

func (cm *ContextManager) getRelevantNPCs(ctx *PlayerContext) []NPCContextInfo {
	var npcs []NPCContextInfo
	
	for _, npcRel := range ctx.NPCStates {
		// Include NPCs the player has interacted with recently
		if time.Since(npcRel.LastInteraction) < 24*time.Hour {
			relationship := cm.determineRelationshipLevel(npcRel.Disposition)
			
			npc := NPCContextInfo{
				ID:           npcRel.NPCID,
				Name:         npcRel.Name,
				Disposition:  npcRel.Disposition,
				Mood:         npcRel.Mood,
				KnownFacts:   npcRel.KnownFacts,
				LastSeen:     cm.formatTimeSince(npcRel.LastInteraction),
				Location:     npcRel.Location,
				Relationship: relationship,
			}
			npcs = append(npcs, npc)
		}
	}

	return npcs
}

func (cm *ContextManager) determinePlayerMood(ctx *PlayerContext) string {
	// Analyze recent actions and outcomes to determine mood
	recentActions := ctx.Actions
	if len(recentActions) == 0 {
		return "curious"
	}

	// Look at last few actions
	startIdx := len(recentActions) - 5
	if startIdx < 0 {
		startIdx = 0
	}

	combatCount := 0
	socialCount := 0
	exploreCount := 0
	successCount := 0

	for i := startIdx; i < len(recentActions); i++ {
		action := recentActions[i]
		
		switch action.Type {
		case "combat", "attack":
			combatCount++
		case "talk", "social":
			socialCount++
		case "explore", "examine":
			exploreCount++
		}
		
		if strings.Contains(strings.ToLower(action.Outcome), "success") {
			successCount++
		}
	}

	totalRecent := len(recentActions) - startIdx
	successRate := float64(successCount) / float64(totalRecent)

	// Determine mood based on activity and success
	if combatCount > socialCount && combatCount > exploreCount {
		if successRate > 0.6 {
			return "confident"
		}
		return "aggressive"
	} else if socialCount > combatCount && socialCount > exploreCount {
		return "diplomatic"
	} else if exploreCount > combatCount {
		return "curious"
	}

	if successRate > 0.7 {
		return "triumphant"
	} else if successRate < 0.3 {
		return "frustrated"
	}

	return "focused"
}

func (cm *ContextManager) formatRecentActions(actions []ActionEvent) string {
	if len(actions) == 0 {
		return "- No recent actions"
	}

	var formatted []string
	for _, action := range actions {
		timeAgo := cm.formatTimeSince(action.Timestamp)
		entry := fmt.Sprintf("- %s ago: %s (%s) -> %s", 
			timeAgo, action.Command, action.Type, action.Outcome)
		formatted = append(formatted, entry)
	}

	return strings.Join(formatted, "\n")
}

func (cm *ContextManager) formatActiveNPCs(npcs []NPCContextInfo) string {
	if len(npcs) == 0 {
		return "- No known NPCs in area"
	}

	var formatted []string
	for _, npc := range npcs {
		entry := fmt.Sprintf("- %s (%s): %s mood, %s relationship (last seen %s)",
			npc.Name, npc.ID, npc.Mood, npc.Relationship, npc.LastSeen)
		if len(npc.KnownFacts) > 0 {
			entry += fmt.Sprintf(" - Knows: %s", strings.Join(npc.KnownFacts, ", "))
		}
		formatted = append(formatted, entry)
	}

	return strings.Join(formatted, "\n")
}

func (cm *ContextManager) formatEquipment(equipment []EquipmentItem) string {
	if len(equipment) == 0 {
		return "No equipment"
	}

	var items []string
	for _, item := range equipment {
		items = append(items, fmt.Sprintf("%s (%s)", item.Name, item.Type))
	}

	return strings.Join(items, ", ")
}

func (cm *ContextManager) formatPreviousLocation(previous string) string {
	if previous == "" {
		return "none"
	}
	return previous
}

func (cm *ContextManager) getReputationDescription(reputation int) string {
	switch {
	case reputation >= 75:
		return "Heroic"
	case reputation >= 50:
		return "Well-regarded"
	case reputation >= 25:
		return "Respected"
	case reputation >= 0:
		return "Neutral"
	case reputation >= -25:
		return "Mistrusted"
	case reputation >= -50:
		return "Disliked"
	default:
		return "Notorious"
	}
}

func (cm *ContextManager) determinePlayerFocus(ctx *PlayerContext) string {
	stats := ctx.SessionStats
	
	if stats.CombatActions > stats.SocialActions && stats.CombatActions > stats.ExploreActions {
		return "Combat-focused"
	} else if stats.SocialActions > stats.CombatActions && stats.SocialActions > stats.ExploreActions {
		return "Social interaction"
	} else if stats.ExploreActions > stats.CombatActions {
		return "Exploration"
	}
	
	return "Balanced gameplay"
}

func (cm *ContextManager) formatWorldContext(worldState map[string]interface{}) string {
	var context []string
	
	if val, ok := worldState["locations_visited"].(int); ok {
		context = append(context, fmt.Sprintf("- Locations explored: %d", val))
	}
	
	if val, ok := worldState["combat_experienced"].(bool); ok && val {
		context = append(context, "- Has combat experience")
	}
	
	if val, ok := worldState["social_active"].(bool); ok && val {
		context = append(context, "- Prefers social interactions")
	}
	
	if len(context) == 0 {
		return "- New to this world"
	}
	
	return strings.Join(context, "\n")
}

func (cm *ContextManager) determineRelationshipLevel(disposition int) string {
	switch {
	case disposition >= 75:
		return "close_friend"
	case disposition >= 50:
		return "friend"
	case disposition >= 25:
		return "ally"
	case disposition >= 0:
		return "acquaintance"
	case disposition >= -25:
		return "stranger"
	case disposition >= -50:
		return "rival"
	default:
		return "enemy"
	}
}

func (cm *ContextManager) formatTimeSince(t time.Time) string {
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "moments"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d min", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hr", hours)
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d days", days)
	}
}

func (cm *ContextManager) determinePlayStyle(ctx *PlayerContext) string {
	stats := ctx.SessionStats
	total := stats.CombatActions + stats.SocialActions + stats.ExploreActions
	
	if total == 0 {
		return "new_player"
	}
	
	combatRatio := float64(stats.CombatActions) / float64(total)
	socialRatio := float64(stats.SocialActions) / float64(total)
	exploreRatio := float64(stats.ExploreActions) / float64(total)
	
	if combatRatio > 0.5 {
		return "combat_focused"
	} else if socialRatio > 0.4 {
		return "roleplay_focused"
	} else if exploreRatio > 0.4 {
		return "exploration_focused"
	}
	
	return "balanced"
}

func (cm *ContextManager) determineExperienceLevel(ctx *PlayerContext) string {
	totalActions := ctx.SessionStats.TotalActions
	sessionTime := time.Since(ctx.StartTime).Minutes()
	
	if totalActions < 10 || sessionTime < 15 {
		return "beginner"
	} else if totalActions < 50 || sessionTime < 60 {
		return "intermediate"
	}
	
	return "experienced"
}

func (cm *ContextManager) getPreferredActivities(ctx *PlayerContext) []string {
	stats := ctx.SessionStats
	activities := []string{}
	
	if stats.CombatActions > 5 {
		activities = append(activities, "combat")
	}
	if stats.SocialActions > 5 {
		activities = append(activities, "social_interaction")
	}
	if stats.ExploreActions > 5 {
		activities = append(activities, "exploration")
	}
	
	if len(activities) == 0 {
		activities = append(activities, "discovering_the_world")
	}
	
	return activities
}

func (cm *ContextManager) getKnownLocations(ctx *PlayerContext) []string {
	locations := make(map[string]bool)
	
	// Add current and previous locations
	locations[ctx.Location.Current] = true
	if ctx.Location.Previous != "" {
		locations[ctx.Location.Previous] = true
	}
	
	// Add locations from history
	for _, visit := range ctx.Location.LocationHistory {
		locations[visit.Location] = true
	}
	
	// Add locations from actions
	for _, action := range ctx.Actions {
		if action.Location != "" {
			locations[action.Location] = true
		}
	}
	
	result := make([]string, 0, len(locations))
	for location := range locations {
		result = append(result, location)
	}
	
	return result
}

func (cm *ContextManager) getEstablishedNPCs(ctx *PlayerContext) []string {
	npcs := make([]string, 0, len(ctx.NPCStates))
	
	for _, npc := range ctx.NPCStates {
		npcs = append(npcs, npc.Name)
	}
	
	return npcs
}

func (cm *ContextManager) getOngoingStorylines(ctx *PlayerContext) []string {
	// This would analyze actions and NPC interactions to identify ongoing storylines
	// For now, return basic storylines based on reputation and interactions
	storylines := []string{}
	
	if ctx.Character.Reputation > 25 {
		storylines = append(storylines, "Building positive reputation in the community")
	} else if ctx.Character.Reputation < -25 {
		storylines = append(storylines, "Dealing with negative reputation consequences")
	}
	
	if len(ctx.NPCStates) > 3 {
		storylines = append(storylines, "Developing relationships with multiple NPCs")
	}
	
	if ctx.SessionStats.CombatActions > 10 {
		storylines = append(storylines, "Engaging in frequent combat encounters")
	}
	
	if len(storylines) == 0 {
		storylines = append(storylines, "Beginning their adventure")
	}
	
	return storylines
}
