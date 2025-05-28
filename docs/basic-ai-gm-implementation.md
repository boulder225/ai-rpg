---
title: "Basic AI GM Core - Implementation Guide"
date: "2025-05-25"
version: "1.0"
---

# Basic AI GM Core Implementation

## Overview

Build autonomous AI Game Master capable of running 30-minute RPG sessions without human intervention. Focus on contextual awareness, personality consistency, and dynamic response generation.

---

## Core Architecture

### AI GM Personality Framework

**Base Personality Matrix:**
```
Helpful vs Challenging: 70/30 (Guides players but creates obstacles)
Mysterious vs Direct: 60/40 (Reveals information strategically)  
Immersive vs Meta: 90/10 (Stays in character, minimal game mechanics exposure)
Reactive vs Proactive: 50/50 (Responds to player actions, introduces events)
```

**Response Categories:**
- **Environmental Descriptions**: Room/location details, atmosphere
- **NPC Interactions**: Dialogue, behavior, reactions  
- **Event Narration**: Combat, skill checks, consequences
- **Quest Guidance**: Hints, objectives, progress updates
- **Emotional Context**: Mood setting, tension building

---

## Prompt Engineering Strategy

### Master System Prompt Structure

**Core Identity:**
```
You are an expert AI Game Master running a fantasy RPG session. Your role:

PERSONALITY: Helpful yet challenging guide who creates immersive experiences
TONE: Descriptive, engaging, appropriate to fantasy setting
GOALS: Player agency, narrative flow, consistent world-building

CRITICAL RULES:
- Always respond in character as the GM
- Maintain world consistency across interactions  
- React contextually to player actions and equipment
- Balance guidance with player discovery
- Generate consequences for player choices
```

**Context Management Template:**
```
CURRENT SESSION STATE:
- Location: [Dynamic based on player position]
- Time: [Track day/night cycle]
- Player Status: [Health, equipment, reputation]
- Active NPCs: [Characters in scene with personality tags]
- Recent Events: [Last 3 significant actions/consequences]

WORLD KNOWLEDGE:
- Core Lore: [Consistent fantasy setting elements]
- NPC Directory: [Character relationships and motivations]
- Location Database: [Places with specific details/secrets]
```

### Response Generation Patterns

**Environmental Awareness:**
- Input: Player action or location change
- Process: Analyze context + current state + personality
- Output: Immersive description with actionable elements

**Example Flow:**
```
Player Input: "/look around"
Context Check: Location = "Thornwick Village Square", Time = "Evening"
Personality Filter: Helpful + Mysterious
Generated Response: "The village square bustles with evening activity. 
Warm light spills from the Rusty Dragon tavern where locals gather. 
You notice a hooded figure watching you from near the well, 
but they quickly look away when you turn their direction."
```

---

## Context Management System

### State Tracking Components

**Player State Object:**
```json
{
  "character": {
    "name": "string",
    "health": "current/max",
    "equipment": ["items worn/carried"],
    "reputation": "number (-100 to 100)"
  },
  "location": {
    "current": "location_id",
    "previous": "location_id", 
    "visit_count": "number"
  },
  "session": {
    "duration": "minutes",
    "actions_taken": "count",
    "last_interaction": "timestamp"
  }
}
```

**World State Tracking:**
```json
{
  "active_npcs": [
    {
      "id": "tavern_keeper_marcus",
      "location": "rusty_dragon_tavern",
      "mood": "cautious", 
      "knows_player": "boolean",
      "last_interaction": "summary"
    }
  ],
  "events": [
    {
      "type": "player_action",
      "description": "entered tavern with bloodied armor",
      "consequences": ["npcs_noticed", "reputation_impact"],
      "timestamp": "session_time"
    }
  ]
}
```

### Memory Management

**Short-term Memory (Current Session):**
- Last 10 player actions and GM responses
- Active NPC states and locations
- Current quest objectives and progress
- Recent environmental changes

**Context Window Optimization:**
- Prioritize recent interactions (higher weight)
- Compress older events into summaries
- Maintain key relationship states
- Track critical world changes

---

## Response Generation Engine

### Decision Tree Framework

**Input Processing:**
1. Parse player command/action
2. Validate against current context
3. Determine response category
4. Apply personality filters
5. Generate contextual response

**Response Categories & Examples:**

**Environmental Description:**
- Trigger: Location change, "/look" command
- Template: "[Atmospheric setting] + [Interactive elements] + [Potential hooks]"
- Example: "Ancient stone walls tower above you, carved with faded runes. A narrow staircase spirals upward into darkness, while a heavy wooden door stands slightly ajar to your left."

**NPC Interaction:**
- Trigger: Player talks to character
- Template: "[NPC greeting based on relationship] + [Current mood/motivation] + [Quest/information hook]"
- Example: "The blacksmith wipes sweat from her brow and eyes your damaged blade. 'That's seen some action. I can fix it, but it'll cost you 15 gold pieces.'"

**Event Narration:**
- Trigger: Combat, skill check, consequence
- Template: "[Action description] + [Dice result integration] + [Outcome with consequences]"
- Example: "You swing your sword at the goblin (rolled 15+3=18). Your blade finds its mark, slicing across its shoulder. The creature stumbles backward, snarling in pain."

### Dynamic Content Generation

**Contextual Hooks:**
- Environmental: Hidden passages, mysterious objects, atmospheric details
- Social: NPC motivations, relationship changes, reputation effects  
- Narrative: Quest opportunities, world events, character development

**Adaptive Difficulty:**
- Monitor player success/failure rates
- Adjust challenge descriptions and outcomes
- Balance guidance vs discovery
- Scale NPC helpfulness to player needs

---

## Integration Patterns

### API Communication Flow

**Request Structure:**
```json
{
  "player_input": "user command or action",
  "session_context": "current game state",
  "world_state": "npc and environment data",
  "response_type": "description/dialogue/narration"
}
```

**Response Format:**
```json
{
  "gm_response": "contextual narrative text",
  "state_updates": "changes to track",
  "suggested_actions": "player options",
  "meta_info": "mechanics/dice results"
}
```

### Error Handling & Fallbacks

**Fallback Response Categories:**
- Generic environmental descriptions
- Standard NPC placeholder dialogue
- Basic consequence acknowledgments
- Session continuity bridging responses

**Quality Assurance Checks:**
- Response length validation (50-200 words)
- Character consistency verification
- World lore accuracy checking
- Appropriate tone maintenance

---

## Testing & Validation Criteria

### Functional Testing

**Core Capabilities:**
- [ ] Maintains consistent GM personality across 30-minute session
- [ ] Responds contextually to player equipment and reputation
- [ ] Generates unique descriptions for repeated location visits
- [ ] Tracks and references previous player actions
- [ ] Balances guidance with player agency

**Performance Benchmarks:**
- Response time: <3 seconds per interaction
- Context accuracy: 90% relevant to current situation  
- Personality consistency: 95% alignment with defined traits
- Session continuity: No contradictions within single session

### User Experience Validation

**Immersion Metrics:**
- Player stays in character without breaking immersion
- Natural conversation flow without awkward AI responses
- Appropriate challenge level maintained throughout session
- Clear progression and consequence feedback

**Success Indicators:**
- Players complete 10+ minute sessions without confusion
- GM responses feel natural and engaging
- World feels consistent and believable
- Player actions have meaningful consequences

---

## Implementation Milestones

### Week 1: Core Personality Foundation
- Define GM personality matrix and response patterns
- Create base system prompt with world knowledge
- Implement basic context tracking
- Test simple interaction scenarios

### Week 2: Contextual World Awareness  
- Add location-based response generation
- Implement equipment and reputation tracking
- Create environmental description templates
- Test location transitions and state persistence

### Week 3: Player Action Recognition
- Build action parsing and consequence system
- Add NPC reaction patterns based on player behavior
- Implement dice roll integration and narration
- Test complex interaction sequences

### Week 4: Session Continuity
- Optimize context window management
- Add session memory and reference capabilities
- Implement fallback systems for error handling
- Conduct full 30-minute session testing

**Success Criteria**: AI GM successfully runs complete demo session with contextual awareness, personality consistency, and player engagement.