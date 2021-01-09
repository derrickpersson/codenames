package codenames

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const wordsPerGame = 25

type GameStage int

const (
	Setup GameStage = iota
	EndSetup
	Explain
	EndExplain
	Gestures
	EndGestures
	OneWord
	EndOneWord
)

func (s GameStage) String() string {
	switch s {
	case Explain:
		return "explain"
	case OneWord:
		return "oneword"
	case Gestures:
		return "gestures"
	case EndSetup:
		return "endsetup"
	case EndExplain:
		return "endexplain"
	case EndOneWord:
		return "endOneWord"
	case EndGestures:
		return "endGestures"
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
		Revealed:  make([]bool, 0),
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
	state.Revealed = make([]bool, 0)
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
	TeamPoints    []TeamPoint  `json:"team_points"`
	CurrentPlayer int          `json:"current_player"`
	RoutingOrder  []TeamPlayer `json:"routing_order"`
	CurrentWord   string       `json:"current_word"`
}

type TeamPlayer struct {
	Team       Team   `json:"team"`
	PlayerName string `json:"player_name"`
}

type TeamPoint struct {
	Team   Team `json:"team"`
	Points int  `json:"points"`
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
	g.UpdatedAt = time.Now()
	g.Round++
	g.getNextPlayer()
	g.GetNextWord(false)
	g.RoundStartedAt = time.Now()
	return true
}

func (g *Game) updateTeamScore(team Team) {
	newTeamPoints := make([]TeamPoint, 0)
	for _, t := range g.TeamPoints {
		if t.Team == team {
			t.Points++
		}
		newTeamPoints = append(newTeamPoints, t)
	}
	g.TeamPoints = append(newTeamPoints)

	g.UpdatedAt = time.Now()
}

func (g *Game) setWinningTeam() {
	var topPoints = 0
	var topTeam = Neutral
	for _, t := range g.TeamPoints {
		if t.Points == topPoints {
			topTeam = Neutral
		} else if t.Points > topPoints {
			topPoints = t.Points
			topTeam = t.Team
		}
	}
	g.WinningTeam = &topTeam
}

func (g *Game) MoveToNextStage() {
	if g.Stage == OneWord {
		g.setWinningTeam()
	} else {
		g.Stage++
		g.GameState.Revealed = make([]bool, len(g.Words))
		g.CurrentWord = ""
	}
	g.UpdatedAt = time.Now()
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
			if value == g.CurrentWord {
				g.GameState.Revealed[idx] = true
			}
		}
	}

	availableWords := g.getAvailableWords()

	if len(availableWords) == 0 {
		g.MoveToNextStage()
	} else {
		idx := rand.Intn(len(availableWords))

		pick := availableWords[idx]

		g.CurrentWord = pick
	}
	g.UpdatedAt = time.Now()
}

func (g *Game) currentTeam() Team {
	if g.Round%2 == 0 {
		return g.StartingTeam
	}
	return g.StartingTeam.Other()
}

func (g *Game) AddWord(word string) error {
	if g.Stage == Setup {
		g.UpdatedAt = time.Now()
		wordIdx := findWordIndex(g.Words, word)
		if wordIdx == -1 {
			g.Words = append(g.Words, word)
			g.GameState.Revealed = append(g.GameState.Revealed, false)
		}
	} else {
		return errors.New("can't add words when past the setup stage")
	}
	return nil
}

func (g *Game) DeleteWord(word string) error {
	if g.Stage == Setup {
		wordIdx := findWordIndex(g.Words, word)
		if wordIdx == -1 {
			return errors.New("Player not found")
		}
		g.Words[len(g.Words)-1], g.Words[wordIdx] = g.Words[wordIdx], g.Words[len(g.Words)-1]
		g.GameState.Revealed[len(g.GameState.Revealed)-1], g.GameState.Revealed[wordIdx] = g.GameState.Revealed[wordIdx], g.GameState.Revealed[len(g.GameState.Revealed)-1]
		g.UpdatedAt = time.Now()
		g.Words = g.Words[:len(g.Words)-1]
		g.GameState.Revealed = g.GameState.Revealed[:len(g.GameState.Revealed)-1]
	}
	return nil
}

func (g *Game) createRoutingOrder(teamPlayers []TeamPlayer) []TeamPlayer {
	turnOrder := make([]TeamPlayer, 0)

	teamRed := make([]TeamPlayer, 0)
	teamBlue := make([]TeamPlayer, 0)

	for idx, tp := range teamPlayers {
		if tp.Team == Red {
			teamRed = append(teamRed, teamPlayers[idx])
		} else if tp.Team == Blue {
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

func (g *Game) UpdatePlayer(oldPlayerName string, team Team, updatedName string) error {
	if err := g.DeletePlayer(oldPlayerName); err != nil {
		return err
	}

	newName := oldPlayerName

	if len(updatedName) > 0 {
		newName = updatedName
	}

	if err := g.AddPlayer(TeamPlayer{PlayerName: newName, Team: team}); err != nil {
		return err
	}
	return nil
}

func (g *Game) AddPlayer(player TeamPlayer) error {
	if g.Stage == Setup {
		// Check for unique name ?
		g.UpdatedAt = time.Now()
		g.TeamPlayers = append(g.TeamPlayers, player)
		g.RoutingOrder = g.createRoutingOrder(g.TeamPlayers)
	} else {
		return errors.New("can't add players when past the setup stage")
	}
	return nil
}

func findWordIndex(s []string, search string) int {
	for idx, item := range s {
		lcItem := strings.ToLower(item)
		lcSearch := strings.ToLower(search)
		if lcItem == lcSearch {
			return idx
		}
	}
	return -1
}

func findPlayerIndex(s []TeamPlayer, search string) int {
	for idx, item := range s {
		if item.PlayerName == search {
			return idx
		}
	}
	return -1
}

func (g *Game) ChangePlayerTeam(name string, team Team) error {
	if err := g.DeletePlayer(name); err != nil {
		return err
	}
	if err := g.AddPlayer(TeamPlayer{PlayerName: name, Team: team}); err != nil {
		return err
	}
	return nil
}

func (g *Game) GetRandomTeam() Team {
	randRnd := rand.New(rand.NewSource(rand.Int63()))
	randomTeam := Team(randRnd.Intn(2)) + Red
	return randomTeam
}

func (g *Game) DeletePlayer(name string) error {
	if g.Stage == Setup {
		playerIdx := findPlayerIndex(g.TeamPlayers, name)
		if playerIdx == -1 {
			return errors.New("Player not found")
		}
		g.TeamPlayers[len(g.TeamPlayers)-1], g.TeamPlayers[playerIdx] = g.TeamPlayers[playerIdx], g.TeamPlayers[len(g.TeamPlayers)-1]
		g.UpdatedAt = time.Now()
		g.TeamPlayers = g.TeamPlayers[:len(g.TeamPlayers)-1]
		g.RoutingOrder = g.createRoutingOrder(g.TeamPlayers)
	} else {
		return errors.New("can't remove players when past the setup stage")
	}
	return nil
}

func (g *Game) getNextPlayer() {
	if g.CurrentPlayer+1 == len(g.RoutingOrder) {
		g.CurrentPlayer = 0
	} else {
		g.CurrentPlayer++
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
		CurrentPlayer:  0, // Circular array reference indx for routingOrder
		RoutingOrder:   make([]TeamPlayer, 0, 0),
		CurrentWord:    ``,
	}

	game.TeamPoints = append(game.TeamPoints, TeamPoint{
		Team:   game.StartingTeam,
		Points: 0,
	})

	game.TeamPoints = append(game.TeamPoints, TeamPoint{
		Team:   game.StartingTeam.Other(),
		Points: 0,
	})

	if opts.RandomWords {
		// Pick the next `wordsPerGame` words from the
		// randomly generated permutation
		perm := seedRnd.Perm(len(state.WordSet))
		permIndex := state.PermIndex
		for _, i := range perm[permIndex : permIndex+wordsPerGame] {
			w := state.WordSet[perm[i]]
			game.Words = append(game.Words, w)
			game.GameState.Revealed = append(game.GameState.Revealed, false)
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
