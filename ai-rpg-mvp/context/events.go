package context

import (
	"log"
	"time"
)

// processEvents processes context events in the background
func (cm *ContextManager) processEvents() {
	defer cm.wg.Done()
	
	for {
		select {
		case event := <-cm.eventQueue:
			cm.processContextEvent(event)
		case <-cm.shutdownCh:
			// Process remaining events before shutdown
			for {
				select {
				case event := <-cm.eventQueue:
					cm.processContextEvent(event)
				default:
					return
				}
			}
		}
	}
}

// processContextEvent processes a single context event
func (cm *ContextManager) processContextEvent(event ContextEvent) {
	ctx, err := cm.GetContext(event.SessionID)
	if err != nil {
		log.Printf("Error getting context for session %s: %v", event.SessionID, err)
		return
	}

	// Add action to history
	ctx.Actions = append(ctx.Actions, event.Event)

	// Trim action history if too long
	if len(ctx.Actions) > cm.maxActions {
		ctx.Actions = ctx.Actions[len(ctx.Actions)-cm.maxActions:]
	}

	// Process action consequences
	cm.processActionConsequences(ctx, event.Event)

	// Update session stats
	cm.updateSessionStats(ctx, event.Event)

	ctx.LastUpdate = time.Now()
	cm.cache.Store(event.SessionID, ctx)
}

// processActionConsequences processes the consequences of a player action
func (cm *ContextManager) processActionConsequences(ctx *PlayerContext, action ActionEvent) {
	for _, consequence := range action.Consequences {
		switch consequence {
		case "reputation_increase":
			change := 5
			if val, ok := action.Metadata["reputation_change"].(int); ok {
				change = val
			}
			ctx.Character.Reputation += change
			
		case "reputation_decrease":
			change := -10
			if val, ok := action.Metadata["reputation_change"].(int); ok {
				change = val
			}
			ctx.Character.Reputation += change
			
		case "health_damage":
			if damage, ok := action.Metadata["damage"].(int); ok {
				ctx.Character.Health.Current -= damage
				if ctx.Character.Health.Current < 0 {
					ctx.Character.Health.Current = 0
				}
			}
			
		case "health_heal":
			if healing, ok := action.Metadata["healing"].(int); ok {
				ctx.Character.Health.Current += healing
				if ctx.Character.Health.Current > ctx.Character.Health.Max {
					ctx.Character.Health.Current = ctx.Character.Health.Max
				}
			}
			
		case "npc_noticed":
			if npcID, ok := action.Metadata["npc_id"].(string); ok {
				if npcName, ok := action.Metadata["npc_name"].(string); ok {
					cm.UpdateNPCRelationship(ctx.SessionID, npcID, npcName, 0, []string{
						"noticed_player_" + action.Type,
					})
				}
			}
			
		case "combat_victory":
			ctx.Character.Reputation += 2
			
		case "combat_defeat":
			ctx.Character.Reputation -= 1
			
		case "quest_completed":
			if reward, ok := action.Metadata["reputation_reward"].(int); ok {
				ctx.Character.Reputation += reward
			}
			
		case "item_gained":
			if itemData, ok := action.Metadata["item"].(map[string]interface{}); ok {
				item := InventoryItem{
					ID:       itemData["id"].(string),
					Name:     itemData["name"].(string),
					Type:     itemData["type"].(string),
					Quantity: 1,
					Value:    0,
					Metadata: make(map[string]interface{}),
				}
				if quantity, ok := itemData["quantity"].(int); ok {
					item.Quantity = quantity
				}
				if value, ok := itemData["value"].(int); ok {
					item.Value = value
				}
				ctx.Character.Inventory = append(ctx.Character.Inventory, item)
			}
			
		case "item_lost":
			if itemID, ok := action.Metadata["item_id"].(string); ok {
				cm.removeItemFromInventory(ctx, itemID)
			}
		}
	}
	
	// Clamp values to valid ranges
	if ctx.Character.Reputation > 100 {
		ctx.Character.Reputation = 100
	} else if ctx.Character.Reputation < -100 {
		ctx.Character.Reputation = -100
	}
}

// updateSessionStats updates session statistics based on action
func (cm *ContextManager) updateSessionStats(ctx *PlayerContext, action ActionEvent) {
	ctx.SessionStats.TotalActions++
	ctx.SessionStats.SessionTime = time.Since(ctx.StartTime).Minutes()
	
	switch action.Type {
	case "combat", "attack", "defend":
		ctx.SessionStats.CombatActions++
	case "talk", "dialogue", "social":
		ctx.SessionStats.SocialActions++
	case "move", "explore", "examine", "look":
		ctx.SessionStats.ExploreActions++
	}
}

// removeItemFromInventory removes an item from player inventory
func (cm *ContextManager) removeItemFromInventory(ctx *PlayerContext, itemID string) {
	for i, item := range ctx.Character.Inventory {
		if item.ID == itemID {
			// Remove item from inventory
			ctx.Character.Inventory = append(ctx.Character.Inventory[:i], ctx.Character.Inventory[i+1:]...)
			break
		}
	}
}

// persistentSaver periodically saves contexts to storage
func (cm *ContextManager) persistentSaver() {
	defer cm.wg.Done()
	
	ticker := time.NewTicker(cm.persistInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			cm.saveAllCachedContexts()
		case <-cm.shutdownCh:
			cm.saveAllCachedContexts()
			return
		}
	}
}

// saveAllCachedContexts saves all cached contexts to storage
func (cm *ContextManager) saveAllCachedContexts() {
	cm.cache.Range(func(key, value interface{}) bool {
		ctx := value.(*PlayerContext)
		if err := cm.storage.SaveContext(ctx); err != nil {
			log.Printf("Error saving context for session %s: %v", ctx.SessionID, err)
		}
		return true
	})
}

// cleanupOldContexts removes old contexts from cache
func (cm *ContextManager) cleanupOldContexts() {
	cutoff := time.Now().Add(-cm.cacheTimeout)
	
	cm.cache.Range(func(key, value interface{}) bool {
		ctx := value.(*PlayerContext)
		if ctx.LastUpdate.Before(cutoff) {
			// Save before removing from cache
			if err := cm.storage.SaveContext(ctx); err != nil {
				log.Printf("Error saving context during cleanup: %v", err)
			}
			cm.cache.Delete(key)
		}
		return true
	})
}

// GetContextMetrics returns metrics about the context manager
func (cm *ContextManager) GetContextMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// Count cached contexts
	cacheCount := 0
	cm.cache.Range(func(key, value interface{}) bool {
		cacheCount++
		return true
	})
	
	metrics["cached_contexts"] = cacheCount
	metrics["event_queue_size"] = len(cm.eventQueue)
	metrics["max_actions"] = cm.maxActions
	metrics["cache_timeout_minutes"] = cm.cacheTimeout.Minutes()
	metrics["persist_interval_minutes"] = cm.persistInterval.Minutes()
	
	return metrics
}

// FlushContext forces immediate save of a specific context
func (cm *ContextManager) FlushContext(sessionID string) error {
	if cached, ok := cm.cache.Load(sessionID); ok {
		ctx := cached.(*PlayerContext)
		return cm.storage.SaveContext(ctx)
	}
	return nil
}

// GetActiveSessions returns list of active session IDs
func (cm *ContextManager) GetActiveSessions() []string {
	var sessions []string
	
	cm.cache.Range(func(key, value interface{}) bool {
		sessionID := key.(string)
		sessions = append(sessions, sessionID)
		return true
	})
	
	return sessions
}

// GetSessionDuration returns how long a session has been active
func (cm *ContextManager) GetSessionDuration(sessionID string) (time.Duration, error) {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return 0, err
	}
	
	return time.Since(ctx.StartTime), nil
}

// IsSessionActive checks if a session is currently active
func (cm *ContextManager) IsSessionActive(sessionID string) bool {
	_, ok := cm.cache.Load(sessionID)
	return ok
}
