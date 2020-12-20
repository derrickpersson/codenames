package codenames

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jbowens/dictionary"
)

var testWords []string

func init() {
	d, err := dictionary.Load("assets/original.txt")
	if err != nil {
		panic(err)
	}
	testWords = d.Words()
}

func BenchmarkGameMarshal(b *testing.B) {
	b.StopTimer()
	d, err := dictionary.Load("assets/original.txt")
	if err != nil {
		b.Fatal(err)
	}
	g := newGame("foo", GameState{
		Seed:     1,
		Round:    0,
		Revealed: make([]bool, 25),
		WordSet:  d.Words(),
	}, GameOptions{RandomWords: true})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err = json.Marshal(g)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGameShuffle(t *testing.T) {
	gamesWithoutRepeats := len(testWords)/25 - 1

	initialState := randomState(testWords)
	currState := initialState

	m := map[string]int{}
	for i := 0; i < gamesWithoutRepeats; i++ {
		g := newGame("foo", currState, GameOptions{RandomWords: true})
		for _, w := range g.Words {
			if prevI, ok := m[w]; ok {
				t.Errorf("Word %q appeared twice, once in game %d and once in game %d.", w, prevI, i)
			}
			m[w] = i
		}
		currState = nextGameState(currState)
	}
}

func TestGameSetUp(t *testing.T) {
	d, err := dictionary.Load("assets/original.txt")
	if err != nil {
		t.Fatal(err)
	}
	initG := newGame("foo", GameState{
		Seed:     1,
		Round:    0,
		Revealed: make([]bool, 25),
		WordSet:  d.Words(),
	}, GameOptions{})
	if initG.Stage != Setup {
		t.Errorf("Failed")
	}
}

func TestGetNextWord(t *testing.T) {
	initG := newGame("foo", GameState{
		Seed:     1,
		Round:    0,
		Revealed: make([]bool, 0),
		WordSet:  make([]string, 0),
	}, GameOptions{})
	initG.AddWord("bar")
	initG.AddWord("foobar")

	initG.GetNextWord(false)
	if initG.currentWord == "" {
		t.Errorf("Current Word not set")
	}
	initG.GetNextWord(true)

	if len(initG.getAvailableWords()) != 1 {
		t.Errorf("Correct word did not get taken out of pool")
	}

	initG.GetNextWord(true)
	if initG.Stage != Explain {
		t.Errorf("Stage not extended")
	}
}

func TestPlayerRouting(t *testing.T) {
	g := newGame("scoring", GameState{
		Seed:     1,
		Round:    0,
		Revealed: make([]bool, 0),
		WordSet:  make([]string, 25),
	}, GameOptions{RandomWords: true})
	playerA := TeamPlayer{team: 1, playerName: "A"}
	player1 := TeamPlayer{team: 0, playerName: "1"}
	playerB := TeamPlayer{team: 1, playerName: "B"}
	player2 := TeamPlayer{team: 0, playerName: "2"}
	player3 := TeamPlayer{team: 0, playerName: "3"}
	g.AddPlayer(playerA)
	g.AddPlayer(player1)

	if g.routingOrder[0] != playerA {
		t.Errorf("Wrong Routing Order")
	}

	if g.routingOrder[1] != player1 {
		t.Errorf("Wrong Routing Order")
	}

	g.AddPlayer(playerB)
	g.AddPlayer(player2)
	if g.routingOrder[2] != playerB {
		fmt.Println("Player: " + g.routingOrder[2].playerName)
		t.Errorf("Wrong Routing Order")
	}

	g.AddPlayer(player3)
	if g.routingOrder[5] != player3 {
		t.Errorf("Wrong Routing Order")
	}

	if g.routingOrder[g.currentPlayer] != playerA {
		t.Errorf("Wrong player in rotation")
	}

	g.getNextPlayer()
	if g.routingOrder[g.currentPlayer] != player1 {
		t.Errorf("Wrong player in rotation")
	}
	g.getNextPlayer()
	if g.routingOrder[g.currentPlayer] != playerB {
		t.Errorf("Wrong player in rotation")
	}
	g.getNextPlayer()
	if g.routingOrder[g.currentPlayer] != player2 {
		t.Errorf("Wrong player in rotation")
	}
	g.getNextPlayer()
	if g.routingOrder[g.currentPlayer] != playerA {
		t.Errorf("Wrong player in rotation")
	}
	g.getNextPlayer()
	if g.routingOrder[g.currentPlayer] != player3 {
		t.Errorf("Wrong player in rotation")
	}

	g.getNextPlayer() // B
	g.getNextPlayer() // 1
	g.getNextPlayer() // A
	g.getNextPlayer() // 2
	g.getNextPlayer() // B
	g.getNextPlayer() // 3
	g.getNextPlayer() // A
	// Get next player properly loops over all players in the round
	if g.routingOrder[g.currentPlayer] != playerA {
		t.Errorf("Wrong player in rotation")
	}
}
