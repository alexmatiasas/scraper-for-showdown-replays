// Package parser provides functions to parse battle log strings into structured data.
package parser

import (
	"strconv"
	"strings"
)

type BattleLog struct {
	ID          string
	Format      string
	Gen         int
	GameType    string
	Rated       bool
	Players     []Player
	Rules       []string
	TeamPreview []TeamPreviewEntry
	Turns       []Turn
	Winner      string
}

type Player struct {
	ID     string
	Name   string
	Rating int
}

type Turn struct {
	Number    int
	Timestamp int64
	Events    []Event
}

type Event struct {
	Type    string
	Pokemon string
	Species string
	Target  string
	Move    string
	HP      string
	Stat    string
	Amount  int
	Detail  string
}

type TeamPreviewEntry struct {
	PlayerID string
	Name     string
	Gender   string
	HasItem  bool
}

func Parse(log string) (*BattleLog, error) {
	lines := strings.Split(log, "\n")
	battle := &BattleLog{}
	currentTurn := -1

	for _, line := range lines {
		if line == "" || line == "|" {
			continue
		}

		// "|event|arg1|arg2" parts[0]="", parts[1]="event", parts[2]="arg1|arg2"
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 2 {
			continue
		}

		eventType := parts[1]
		args := ""

		if len(parts) > 2 {
			args = parts[2]
		}

		switch eventType {
		// Metadata
		case "player":
			battle.addPlayer(args)
		case "gen":
			battle.Gen = parseGen(args)
		case "gametype":
			battle.GameType = args
		case "rated":
			battle.Rated = true
		case "rule":
			battle.Rules = append(battle.Rules, args)
		case "poke":
			battle.addTeamPreview(args)

			// Turn control
		case "turn":
			turnNum := parseTurnNumber(args)
			currentTurn = turnNum - 1
			battle.startTurn(turnNum)
		case "t:":
			battle.updateTimestamp(currentTurn, args)

			// Battle actions
		case "switch", "drag":
			battle.addEvent(currentTurn, parseSwitch(eventType, args))

		case "move":
			battle.addEvent(currentTurn, parseMove(args))
		case "-damage", "-heal":
			battle.addEvent(currentTurn, parseDamageHeal(eventType, args))
		case "-boost", "-unboost":
			battle.addEvent(currentTurn, parseBoost(eventType, args))
		case "-status":
			battle.addEvent(currentTurn, parseStatus(args))
		case "-ability":
			battle.addEvent(currentTurn, parseAbility(args))
		case "-start", "-end":
			battle.addEvent(currentTurn, parseVolatile(eventType, args))
		case "-terastallize":
			battle.addEvent(currentTurn, parseTerastallize(args))
		case "faint":
			battle.addEvent(currentTurn, parseFaint(args))
		case "win":
			battle.Winner = args
		}

	}

	return battle, nil
}

// --- Auxiliar methods of the BattleLog ---
func (b *BattleLog) addPlayer(args string) {
	// "|player|p1|andjelicpwnsu|wallace|1320"
	parts := strings.Split(args, "|")
	if len(parts) < 4 {
		b.Players = append(b.Players, Player{
			ID:     "",
			Name:   "",
			Rating: -1,
		})
		// TODO: add warning or error to communicate
		return
	}
	rating, _ := strconv.Atoi(parts[3])
	b.Players = append(b.Players, Player{
		ID:     parts[0],
		Name:   parts[1],
		Rating: rating,
	})
}

func (b *BattleLog) addTeamPreview(args string) {
	// "|poke|p1|Hatterene, F|"
	parts := strings.Split(args, "|")
	if len(parts) < 3 {
		b.TeamPreview = append(b.TeamPreview, TeamPreviewEntry{
			PlayerID: "",
			Name:     "",
			Gender:   "",
			HasItem:  false,
		})
		// TODO: add warning or error
		return
	}
	gender := ""
	name := parts[1]
	if strings.Contains(name, ", ") {
		nameParts := strings.SplitN(name, ", ", 2)
		name = nameParts[0]
		gender = nameParts[1]
	}
	b.TeamPreview = append(b.TeamPreview, TeamPreviewEntry{
		PlayerID: parts[0],
		Name:     name,
		Gender:   gender,
		HasItem:  parts[2] != "",
	})
}

func (b *BattleLog) startTurn(number int) {
	b.Turns = append(b.Turns, Turn{
		Number: number,
		Events: []Event{},
	})
}

func (b *BattleLog) updateTimestamp(currentTurn int, args string) {
	ts, _ := strconv.ParseInt(args, 10, 64)
	if currentTurn >= 0 && currentTurn < len(b.Turns) {
		b.Turns[currentTurn].Timestamp = ts
	}
}

func (b *BattleLog) addEvent(turnIndex int, event Event) {
	if turnIndex >= 0 && turnIndex < len(b.Turns) {
		b.Turns[turnIndex].Events = append(b.Turns[turnIndex].Events, event)
	}
}

// --- Parse functions ---

func parseGen(s string) int {
	//  "|gen|9"
	gen, _ := strconv.Atoi(s)
	return gen
}

func parseTurnNumber(s string) int {
	// "|turn|1"
	n, _ := strconv.Atoi(s)
	return n // Turns as they came
}

func parseSwitch(eventType, args string) Event {
	// "|switch|p1a:Pecharunt|Pecharunt|100/100"
	parts := strings.Split(args, "|")
	// validation in case args it's wrong, less args.
	if len(parts) < 3 {
		return Event{Type: eventType}
	}
	return Event{
		Type:    eventType,
		Pokemon: parts[0],
		Species: parts[1], // species name
		HP:      parts[2],
	}
}

func parseMove(args string) Event {
	// "|move|p2a: Kyurem|Ice Beam|p1a: Hatterene"
	parts := strings.Split(args, "|")
	if len(parts) < 3 {
		return Event{Type: "move"}
	}
	return Event{
		Type:    "move",
		Pokemon: parts[0],
		Move:    parts[1],
		Target:  parts[2],
	}
}

func parseDamageHeal(eventType, args string) Event {
	// "|-damage|p1a: Hatterene|0 fnt"
	parts := strings.Split(args, "|")
	if len(parts) < 2 {
		return Event{Type: eventType}
	}
	return Event{
		Type:    eventType,
		Pokemon: parts[0],
		HP:      parts[1],
	}
}

func parseBoost(eventType, args string) Event {
	// "|-boost|p1a: Iron Moth|spa|1"
	parts := strings.Split(args, "|")
	if len(parts) < 3 {
		return Event{Type: eventType}
	}
	amount, _ := strconv.Atoi(parts[2])
	return Event{
		Type:    eventType,
		Pokemon: parts[0],
		Stat:    parts[1],
		Amount:  amount,
	}
}

func parseStatus(args string) Event {
	// "|-status|p1a: Gliscor|tox"
	parts := strings.Split(args, "|")
	if len(parts) < 2 {
		return Event{Type: "status"}
	}
	return Event{
		Type:    "status",
		Pokemon: parts[0],
		Detail:  parts[1],
	}
}

func parseAbility(args string) Event {
	// "|-ability|p2a: Kyurem|Pressure"
	parts := strings.Split(args, "|")
	if len(parts) < 2 {
		return Event{Type: "ability"}
	}
	return Event{
		Type:    "ability",
		Pokemon: parts[0],
		Detail:  parts[1],
	}
}

func parseVolatile(eventType, args string) Event {
	// "|-start|p2a: Kingambit|fallen1|[silent]"
	// TODO: tags [silent], [still] are lost, we need to parse them.
	parts := strings.Split(args, "|")
	if len(parts) < 3 {
		return Event{Type: eventType}
	}
	return Event{
		Type:    eventType,
		Pokemon: parts[0],
		Detail:  parts[1],
	}
}

func parseTerastallize(args string) Event {
	// "|-terastallize|p2a: Kyurem|Ice"
	parts := strings.Split(args, "|")
	if len(parts) < 2 {
		return Event{Type: "terastallize"}
	}
	return Event{
		Type:    "terastallize",
		Pokemon: parts[0],
		Detail:  parts[1],
	}
}

func parseFaint(args string) Event {
	// "|faint|p1a: Hatterene"
	return Event{
		Type:    "faint",
		Pokemon: args,
	}
}
