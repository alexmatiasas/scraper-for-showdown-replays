package parser

import (
	"os"
	"testing"
)

func loadTestLog(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/gen9ou.log")
	if err != nil {
		t.Fatalf("failed to load test log: %v", err)
	}
	// This needs to be casted trough a for loop
	return string(data)
}

// --- Main Test: TestParse_BattleLog ---

func TestParse_BattleLog(t *testing.T) {
	log := loadTestLog(t)
	battle, err := Parse(log)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err) // Fatalf() stops the test
	}

	if battle.Gen != 9 {
		t.Errorf("Gen = %d, want 9", battle.Gen) // Errorf() continues test
	}
	if battle.GameType != "singles" {
		t.Errorf("GameType = %q, want %q", battle.GameType, "singles")
	}
	if !battle.Rated {
		t.Error("Rated = false, want true")
	}
	if battle.Winner != "Alice" {
		t.Errorf("Winner = %q, want %q", battle.Winner, "Alice")
	}
}

// --- Parse Players Test ---

func TestParse_Players(t *testing.T) {
	log := loadTestLog(t)
	battle, _ := Parse(log)

	if len(battle.Players) != 2 {
		t.Fatalf("len(Players) = %d, want 2", len(battle.Players))
	}

	// Player 1: with rating
	if battle.Players[0].ID != "p1" {
		t.Errorf("Players[0].ID = %q, want %q", battle.Players[0].ID, "p1")
	}
	if battle.Players[0].Name != "Alice" {
		t.Errorf("Players[0].Name = %q, want %q", battle.Players[0].Name, "Alice")
	}
	if battle.Players[0].Rating != 1500 {
		t.Errorf("Players[0].Rating = %d, want 1500", battle.Players[0].Rating)
	}

	// Player 2: with no rating (|player|p2|Bob|trainer|)
	// Parser reads parts[3] which is "" → Rating = 0
	if battle.Players[1].Rating != 0 {
		t.Errorf("Players[1].Rating = %d, want 0", battle.Players[1].Rating)
	}
}

// --- Rules Test ---

func TestParse_Rules(t *testing.T) {
	log := loadTestLog(t)
	battle, _ := Parse(log)

	if len(battle.Rules) != 1 {
		t.Fatalf("len(Rules) = %d, want 1", len(battle.Rules))
	}
	expected := "HP Percentage Mod: HP is shown in percentages"
	if battle.Rules[0] != expected {
		t.Errorf("Rules[0] = %q, want %q", battle.Rules[0], expected)
	}
}

// --- Team Preview Test ---

func TestParse_TeamPreview(t *testing.T) {
	log := loadTestLog(t)
	battle, _ := Parse(log)

	if len(battle.TeamPreview) != 2 {
		t.Fatalf("len(TeamPreview) = %d, want 2", len(battle.TeamPreview))
	}

	// Garchomp: sin gender, sin item
	garchomp := battle.TeamPreview[0]
	if garchomp.PlayerID != "p1" {
		t.Errorf("TeamPreview[0].PlayerID = %q, want %q", garchomp.PlayerID, "p1")
	}
	if garchomp.Name != "Garchomp" {
		t.Errorf("TeamPreview[0].Name = %q, want %q", garchomp.Name, "Garchomp")
	}
	if garchomp.Gender != "M" {
		t.Errorf("TeamPreview[0].Gender = %q, want empty", garchomp.Gender)
	}
	if garchomp.HasItem {
		t.Error("TeamPreview[0].HasItem = true, want false")
	}

	// Flutter Mane: con gender, con item
	flutter := battle.TeamPreview[1]
	if flutter.Name != "Flutter Mane" {
		t.Errorf("TeamPreview[1].Name = %q, want %q", flutter.Name, "Flutter Mane")
	}
	if flutter.Gender != "F" {
		t.Errorf("TeamPreview[1].Gender = %q, want %q", flutter.Gender, "F")
	}
	if !flutter.HasItem {
		t.Error("TeamPreview[1].HasItem = false, want true")
	}
}

// --- Turns Test ---

func TestParse_Turns(t *testing.T) {
	log := loadTestLog(t)
	battle, _ := Parse(log)

	if len(battle.Turns) != 3 {
		t.Fatalf("len(Turns) = %d, want 3", len(battle.Turns))
	}

	// Turn 1: number=1 (0-indexed = 0), timestamp
	turn1 := battle.Turns[0]
	if turn1.Number != 1 {
		t.Errorf("Turns[0].Number = %d, want 1", turn1.Number)
	}
	if turn1.Timestamp != 1782964964 {
		t.Errorf("Turns[0].Timestamp = %d, want 1782964964", turn1.Timestamp)
	}

	// Turn 3: more events
	turn3 := battle.Turns[2]
	if turn3.Number != 3 {
		t.Errorf("Turns[2].Number = %d, want 3", turn3.Number)
	}
}

// --- Events by turn Test ---

func TestParse_Events(t *testing.T) {
	log := loadTestLog(t)
	battle, _ := Parse(log)

	// Turn 1: switch + switch + move + -damage + faint = 5 events
	// Note: |-supereffective| NOT in the switch of Parser(), ignored.
	turn1 := battle.Turns[0]
	if len(turn1.Events) != 5 {
		t.Fatalf("Turn 1: len(Events) = %d, want 5", len(turn1.Events))
	}

	// [0] switch
	if turn1.Events[0].Type != "switch" {
		t.Errorf("Turn 1, Event[0].Type = %q, want %q", turn1.Events[0].Type, "switch")
	}
	if turn1.Events[0].Pokemon != "p1a: Garchomp" {
		t.Errorf("Turn 1, Event[0].Pokemon = %q", turn1.Events[0].Pokemon)
	}

	// [1] switch
	if turn1.Events[1].Type != "switch" {
		t.Errorf("Turn 1, Event[1].Type = %q, want %q", turn1.Events[1].Type, "switch")
	}
	if turn1.Events[1].Pokemon != "p2a: Flutter Mane" {
		t.Errorf("Turn 1, Event[1].Pokemon = %q", turn1.Events[1].Pokemon)
	}

	// [2] move
	if turn1.Events[2].Type != "move" {
		t.Errorf("Turn 1, Event[2].Type = %q, want %q", turn1.Events[2].Type, "move")
	}
	if turn1.Events[2].Move != "Earthquake" {
		t.Errorf("Turn 1, Event[2].Move = %q, want %q", turn1.Events[2].Move, "Earthquake")
	}

	// [3] -damage
	if turn1.Events[3].Type != "-damage" {
		t.Errorf("Turn 1, Event[3].Type = %q, want %q", turn1.Events[3].Type, "-damage")
	}

	// [4] faint
	if turn1.Events[4].Type != "faint" {
		t.Errorf("Turn 1, Event[4].Type = %q, want %q", turn1.Events[4].Type, "faint")
	}

	// Turn 3: move + -damage + -status + -ability + -terastallize + move + -damage + faint = 8 events
	turn3 := battle.Turns[2]
	if len(turn3.Events) != 8 {
		t.Fatalf("Turn 3: len(Events) = %d, want 8", len(turn3.Events))
	}

	// [0] move
	if turn3.Events[0].Type != "move" {
		t.Errorf("Turn 3, Event[0].Type = %q, want %q", turn3.Events[0].Type, "move")
	}
	if turn3.Events[0].Move != "Sucker Punch" {
		t.Errorf("Turn 3, Event[0].Move = %q, want %q", turn3.Events[0].Move, "Sucker Punch")
	}

	// [2] status (parser stores "status", not "-status")
	statusEvent := turn3.Events[2]
	if statusEvent.Type != "status" {
		t.Errorf("Turn 3, Event[2].Type = %q, want %q", statusEvent.Type, "status")
	}
	if statusEvent.Detail != "brn" {
		t.Errorf("Turn 3, Event[2].Detail = %q, want %q", statusEvent.Detail, "brn")
	}

	// [4] terastallize (parser stores "terastallize", not "-terastallize")
	teraEvent := turn3.Events[4]
	if teraEvent.Type != "terastallize" {
		t.Errorf("Turn 3, Event[4].Type = %q, want %q", teraEvent.Type, "terastallize")
	}
	if teraEvent.Detail != "Steel" {
		t.Errorf("Turn 3, Event[4].Detail = %q, want %q", teraEvent.Detail, "Steel")
	}
}

// --- Unit Tests of helpers ---

func TestAddPlayer_Valid(t *testing.T) {
	b := &BattleLog{}
	b.addPlayer("p1|Alice|trainer|1500")

	if len(b.Players) != 1 {
		t.Fatalf("len(Players) = %d, want 1", len(b.Players))
	}
	if b.Players[0].ID != "p1" {
		t.Errorf("ID = %q, want %q", b.Players[0].ID, "p1")
	}
	if b.Players[0].Name != "Alice" {
		t.Errorf("Name = %q, want %q", b.Players[0].Name, "Alice")
	}
	if b.Players[0].Rating != 1500 {
		t.Errorf("Rating = %d, want 1500", b.Players[0].Rating)
	}
}

func TestAddPlayer_Invalid(t *testing.T) {
	b := &BattleLog{}
	b.addPlayer("p1|Alice")

	if len(b.Players) != 1 {
		t.Fatalf("len(Players) = %d, want 1", len(b.Players))
	}
	if b.Players[0].Rating != -1 {
		t.Errorf("Rating = %d, want -1", b.Players[0].Rating)
	}
}

func TestAddTeamPreview_Valid(t *testing.T) {
	b := &BattleLog{}
	b.addTeamPreview("p1|Garchomp, M|")

	if len(b.TeamPreview) != 1 {
		t.Fatalf("len(TeamPreview) = %d, want 1", len(b.TeamPreview))
	}
	if b.TeamPreview[0].Name != "Garchomp" {
		t.Errorf("Name = %q, want %q", b.TeamPreview[0].Name, "Garchomp")
	}
	if b.TeamPreview[0].Gender != "M" {
		t.Errorf("Gender = %q, want %q", b.TeamPreview[0].Gender, "M")
	}
	if b.TeamPreview[0].HasItem {
		t.Error("HasItem = true, want false")
	}
}

func TestAddTeamPreview_WithItem(t *testing.T) {
	b := &BattleLog{}
	b.addTeamPreview("p2|Flutter Mane, F|Booster Energy")

	if b.TeamPreview[0].HasItem != true {
		t.Error("HasItem = false, want true")
	}
}

func TestAddTeamPreview_Invalid(t *testing.T) {
	b := &BattleLog{}
	b.addTeamPreview("p1|Garchomp")

	if len(b.TeamPreview) != 1 {
		t.Fatalf("len(TeamPreview) = %d, want 1", len(b.TeamPreview))
	}
	if b.TeamPreview[0].PlayerID != "" {
		t.Errorf("PlayerID = %q, want empty", b.TeamPreview[0].PlayerID)
	}
}

func TestParseSwitch_Valid(t *testing.T) {
	event := parseSwitch("switch", "p1a: Garchomp|Garchomp, M|100/100")
	if event.Pokemon != "p1a: Garchomp" {
		t.Errorf("Pokemon = %q", event.Pokemon)
	}
	if event.HP != "100/100" {
		t.Errorf("HP = %q", event.HP)
	}
}

func TestParseSwitch_Invalid(t *testing.T) {
	event := parseSwitch("switch", "p1a: Garchomp")
	if event.Pokemon != "" {
		t.Errorf("Pokemon = %q, want empty", event.Pokemon)
	}
}

func TestParseMove_Valid(t *testing.T) {
	event := parseMove("p1a: Garchomp|Earthquake|p2a: Flutter Mane")
	if event.Move != "Earthquake" {
		t.Errorf("Move = %q, want %q", event.Move, "Earthquake")
	}
	if event.Target != "p2a: Flutter Mane" {
		t.Errorf("Target = %q", event.Target)
	}
}

func TestParseMove_Invalid(t *testing.T) {
	event := parseMove("p1a: Garchomp")
	if event.Move != "" {
		t.Errorf("Move = %q, want empty", event.Move)
	}
}

func TestParseBoost_Valid(t *testing.T) {
	event := parseBoost("-boost", "p1a: Garchomp|atk|2")
	if event.Stat != "atk" {
		t.Errorf("Stat = %q, want %q", event.Stat, "atk")
	}
	if event.Amount != 2 {
		t.Errorf("Amount = %d, want 2", event.Amount)
	}
}

func TestParseBoost_Invalid(t *testing.T) {
	event := parseBoost("-boost", "p1a: Garchomp")
	if event.Stat != "" {
		t.Errorf("Stat = %q, want empty", event.Stat)
	}
}

func TestParseGen(t *testing.T) {
	if gen := parseGen("9"); gen != 9 {
		t.Errorf("parseGen(\"9\") = %d, want 9", gen)
	}
	if gen := parseGen(""); gen != 0 {
		t.Errorf("parseGen(\"\") = %d, want 0", gen)
	}
}

func TestParseTurnNumber(t *testing.T) {
	// parseTurnNumber hace n - 1 para 0-indexed
	if n := parseTurnNumber("1"); n != 1 {
		t.Errorf("parseTurnNumber(\"1\") = %d, want 0", n)
	}
	if n := parseTurnNumber("5"); n != 5 {
		t.Errorf("parseTurnNumber(\"5\") = %d, want 4", n)
	}
}

func TestEmptyLog(t *testing.T) {
	battle, err := Parse("")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if battle.Gen != 0 {
		t.Errorf("Gen = %d, want 0", battle.Gen)
	}
	if len(battle.Players) != 0 {
		t.Errorf("len(Players) = %d, want 0", len(battle.Players))
	}
	if len(battle.Turns) != 0 {
		t.Errorf("len(Turns) = %d, want 0", len(battle.Turns))
	}
}

// --- parseDamageHeal Test ---

func TestParseDamageHeal(t *testing.T) {
	event := parseDamageHeal("-damage", "p1a: Garchomp|80/100")
	if event.Pokemon != "p1a: Garchomp" {
		t.Errorf("Pokemon = %q", event.Pokemon)
	}
	if event.HP != "80/100" {
		t.Errorf("HP = %q", event.HP)
	}
}

func TestParseDamageHeal_Invalid(t *testing.T) {
	event := parseDamageHeal("-damage", "")
	if event.Pokemon != "" {
		t.Errorf("Pokemon = %q, want empty", event.Pokemon)
	}
}
