package codenames

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const wordsPerGame = 25

type GameStage int

const (
	Setup GameStage = iota
	Explain
	OneWord
	Gestures
)

func (s GameStage) String() string {
	switch s {
	case Explain:
		return "explain"
	case OneWord:
		return "oneword"
	case Gestures:
		return "gestures"
	default:
		return "setup"
	}
}

type Team int

const (
	Red  Team = iota // 0
	Blue             // 1
	Neutral
)

func (t Team) String() string {
	switch t {
	case Red:
		return "red"
	case Blue:
		return "blue"
	default:
		return "neutral"
	}
}

func (t Team) Other() Team {
	if t == Red {
		return Blue
	}
	if t == Blue {
		return Red
	}
	return t
}

func (t *Team) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	switch s {
	case "red":
		*t = Red
	case "blue":
		*t = Blue
	default:
		*t = Neutral
	}
	return nil
}

func (t Team) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t Team) Repeat(n int) []Team {
	s := make([]Team, n)
	for i := 0; i < n; i++ {
		s[i] = t
	}
	return s
}

// GameState encapsulates enough data to reconstruct
// a Game's state. It's used to recreate games after
// a process restart.
type GameState struct {
	Seed      int64    `json:"seed"`
	PermIndex int      `json:"perm_index"`
	Round     int      `json:"round"`
	Revealed  []bool   `json:"revealed"`
	WordSet   []string `json:"word_set"`
}

func (gs GameState) anyRevealed() bool {
	var revealed bool
	for _, r := range gs.Revealed {
		revealed = revealed || r
	}
	return revealed
}

func randomState(words []string) GameState {
	return GameState{
		Seed:      rand.Int63(),
		PermIndex: 0,
		Round:     0,
		Revealed:  make([]bool, wordsPerGame),
		WordSet:   words,
	}
}

// nextGameState returns a new GameState for the next game.
func nextGameState(state GameState) GameState {
	state.PermIndex = state.PermIndex + wordsPerGame
	if state.PermIndex+wordsPerGame >= len(state.WordSet) {
		state.Seed = rand.Int63()
		state.PermIndex = 0
	}
	state.Revealed = make([]bool, wordsPerGame)
	state.Round = 0
	return state
}

type Game struct {
	GameState
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	StartingTeam   Team      `json:"starting_team"`
	WinningTeam    *Team     `json:"winning_team,omitempty"`
	Words          []string  `json:"words"`
	Layout         []Team    `json:"layout"`
	RoundStartedAt time.Time `json:"round_started_at,omitempty"`
	GameOptions
	TeamPlayers   []TeamPlayer `json:"team_players,omitempty"`
	Stage         GameStage    `json:"stage"`
	TeamPoints    []TeamPoint  `json:"team_points,omitempty"`
	currentPlayer int          `json:"current_player"`
	routingOrder  []TeamPlayer `json:"routing_order"`
	currentWord   string       `json:"current_word"`
}

type TeamPlayer struct {
	team       Team
	playerName string
}

type TeamPoint struct {
	team   Team
	points int
}

type GameOptions struct {
	TimerDurationMS int64 `json:"timer_duration_ms,omitempty"`
	EnforceTimer    bool  `json:"enforce_timer,omitempty"`
	RandomWords     bool  `json:"random_words,omitempty"`
}

func (g *Game) StateID() string {
	return fmt.Sprintf("%019d", g.UpdatedAt.UnixNano())
}

func (g *Game) checkWinningCondition() {
	if g.WinningTeam != nil {
		return
	}
	var redRemaining, blueRemaining bool
	for i, t := range g.Layout {
		if g.Revealed[i] {
			continue
		}
		switch t {
		case Red:
			redRemaining = true
		case Blue:
			blueRemaining = true
		}
	}
	if !redRemaining {
		winners := Red
		g.WinningTeam = &winners
	}
	if !blueRemaining {
		winners := Blue
		g.WinningTeam = &winners
	}
}

func (g *Game) NextTurn(currentTurn int) bool {
	if g.WinningTeam != nil {
		return false
	}
	// TODO: remove currentTurn != 0 once we can be sure all
	// clients are running up-to-date versions of the frontend.
	if g.Round != currentTurn && currentTurn != 0 {
		return false
	}
	g.UpdatedAt = time.Now()
	g.Round++
	g.getNextPlayer()
	g.RoundStartedAt = time.Now()
	return true
}

func (g *Game) updateTeamScore(team Team) {
	for _, t := range g.TeamPoints {
		if t.team == team {
			t.points++
		}
	}
}

func (g *Game) moveToNextStage() {
	g.Stage++
	g.GameState.Revealed = make([]bool, len(g.Words))
}

func (g *Game) getAvailableWords() []string {
	availableWords := make([]string, 0)
	for idx, item := range g.Words {
		if !g.GameState.Revealed[idx] {
			availableWords = append(availableWords, item)
		}
	}
	return availableWords
}

func (g *Game) GetNextWord(correct bool) {
	if correct {
		g.updateTeamScore(g.currentTeam())
		for idx, value := range g.Words {
			if value == g.currentWord {
				g.GameState.Revealed[idx] = true
			}
		}
	}

	availableWords := g.getAvailableWords()

	if len(availableWords) == 0 {
		g.moveToNextStage()
	} else {
		idx := rand.Intn(len(availableWords))

		pick := availableWords[idx]

		g.currentWord = pick
	}
	g.UpdatedAt = time.Now()
}

// func (g *Game) Guess(idx int) error {
// 	if idx > len(g.Layout) || idx < 0 {
// 		return fmt.Errorf("index %d is invalid", idx)
// 	}
// 	if g.Revealed[idx] {
// 		return errors.New("cell has already been revealed")
// 	}
// 	g.UpdatedAt = time.Now()
// 	g.Revealed[idx] = true

// 	g.checkWinningCondition()
// 	if g.Layout[idx] != g.currentTeam() {
// 		g.Round = g.Round + 1
// 		g.RoundStartedAt = time.Now()
// 	}
// 	return nil
// }

func (g *Game) currentTeam() Team {
	if g.Round%2 == 0 {
		return g.StartingTeam
	}
	return g.StartingTeam.Other()
}

func (g *Game) AddWord(word string) error {
	if g.Stage == Setup {
		g.Words = append(g.Words, word)
		g.GameState.Revealed = append(g.GameState.Revealed, false)
	} else {
		return errors.New("can't add words when past the setup stage")
	}
	return nil
}

func (g *Game) createRoutingOrder(teamPlayers []TeamPlayer) []TeamPlayer {
	turnOrder := make([]TeamPlayer, 0)

	teamRed := make([]TeamPlayer, 0)
	teamBlue := make([]TeamPlayer, 0)

	for idx, tp := range teamPlayers {
		if tp.team == Red {
			teamRed = append(teamRed, teamPlayers[idx])
		} else if tp.team == Blue {
			teamBlue = append(teamBlue, teamPlayers[idx])
		}
	}

	rotationLength := len(teamRed) * len(teamBlue) * 2
	count := g.StartingTeam
	blueCount := 0
	redCount := 0

	for len(turnOrder) < rotationLength {
		if count%2 == 0 {
			if len(teamRed) > 0 {
				redIdx := redCount % len(teamRed)
				turnOrder = append(turnOrder, teamRed[redIdx])
				redCount++
			}
		} else {
			if len(teamBlue) > 0 {
				blueIdx := blueCount % len(teamBlue)
				turnOrder = append(turnOrder, teamBlue[blueIdx])
				blueCount++
			}
		}
		count++
	}
	return turnOrder
}

func (g *Game) AddPlayer(player TeamPlayer) error {
	if g.Stage == Setup {
		// Check for unique name ?
		g.TeamPlayers = append(g.TeamPlayers, player)
		g.routingOrder = g.createRoutingOrder(g.TeamPlayers)
	} else {
		return errors.New("can't add players when past the setup stage")
	}
	return nil
}

func findPlayerIndex(s []TeamPlayer, search string) int {
	for idx, item := range s {
		if item.playerName == search {
			return idx
		}
	}
	return -1
}

func (g *Game) RemovePlayer(name string) error {
	if g.Stage == Setup {
		playerIdx := findPlayerIndex(g.TeamPlayers, name)
		if playerIdx == -1 {
			return errors.New("Player not found")
		}
		g.TeamPlayers[len(g.TeamPlayers)-1], g.TeamPlayers[playerIdx] = g.TeamPlayers[playerIdx], g.TeamPlayers[len(g.TeamPlayers)-1]
		g.TeamPlayers = g.TeamPlayers[:len(g.TeamPlayers)-1]
		g.routingOrder = g.createRoutingOrder(g.TeamPlayers)
	} else {
		return errors.New("can't remove players when past the setup stage")
	}
	return nil
}

func (g *Game) getNextPlayer() {
	if g.currentPlayer+1 == len(g.routingOrder) {
		g.currentPlayer = 0
	} else {
		g.currentPlayer++
	}
}

func newGame(id string, state GameState, opts GameOptions) *Game {
	// consistent randomness across games with the same seed
	seedRnd := rand.New(rand.NewSource(state.Seed))
	// distinct randomness across games with same seed
	randRnd := rand.New(rand.NewSource(state.Seed * int64(state.PermIndex+1)))

	game := &Game{
		ID:             id,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		StartingTeam:   Team(randRnd.Intn(2)) + Red,
		Words:          make([]string, 0),
		GameState:      state,
		RoundStartedAt: time.Now(),
		GameOptions:    opts,
		TeamPlayers:    make([]TeamPlayer, 0, 0),
		Stage:          Setup,
		TeamPoints:     make([]TeamPoint, 0, 2),
		currentPlayer:  0, // Circular array reference indx for routingOrder
		routingOrder:   make([]TeamPlayer, 0, 0),
		currentWord:    ``,
	}

	if opts.RandomWords {
		// Pick the next `wordsPerGame` words from the
		// randomly generated permutation
		perm := seedRnd.Perm(len(state.WordSet))
		permIndex := state.PermIndex
		for _, i := range perm[permIndex : permIndex+wordsPerGame] {
			w := state.WordSet[perm[i]]
			game.Words = append(game.Words, w)
		}
	}
	return game
}

func shuffle(rnd *rand.Rand, teamAssignments []Team) {
	for i := range teamAssignments {
		j := rnd.Intn(i + 1)
		teamAssignments[i], teamAssignments[j] = teamAssignments[j], teamAssignments[i]
	}
}
