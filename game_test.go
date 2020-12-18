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

func TestGameScoring(t *testing.T) {
	g := newGame("scoring", GameState{
		Seed:     1,
		Round:    0,
		Revealed: make([]bool, 0),
		WordSet:  make([]string, 0),
	}, GameOptions{})

	fmt.Println("CurrentTeam: " + g.StartingTeam.String())
}
