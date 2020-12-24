import * as React from 'react';
import axios from 'axios';
import { Settings, SettingsButton, SettingsPanel } from '~/ui/settings';
import Timer from '~/ui/timer';
import GameSetup from '~/ui/game_setup';
import { IGame } from '~/ui/models';
import GameTurn from '~/ui/game_turn';
import StageEnd from './stage_end';

type GameMode = 'game' | 'spymaster';

interface Props {
  gameID: string;
}

interface State {
  game: IGame;
  mounted: boolean;
  mode: GameMode;
}

const defaultFavicon =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAA8SURBVHgB7dHBDQAgCAPA1oVkBWdzPR84kW4AD0LCg36bXJqUcLL2eVY/EEwDFQBeEfPnqUpkLmigAvABK38Grs5TfaMAAAAASUVORK5CYII=';
const blueTurnFavicon =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAmSURBVHgB7cxBAQAABATBo5ls6ulEiPt47ASYqJ6VIWUiICD4Ehyi7wKv/xtOewAAAABJRU5ErkJggg==';
const redTurnFavicon =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAmSURBVHgB7cwxAQAACMOwgaL5d4EiELGHoxGQGnsVaIUICAi+BAci2gJQFUhklQAAAABJRU5ErkJggg==';
export class Game extends React.Component<Props, State> {
  private state: State;
  private props: Props;
  constructor(props) {
    super(props);
    this.state = {
      game: null,
      mounted: true,
      mode: 'game',
    };
    this.handleAddWord.bind(this);
    this.handleChangePlayerTeam.bind(this);
    this.handleNextStage.bind(this);
    this.handleRemoveWord.bind(this);
    this.handleRemovePlayer.bind(this);
  }

  public componentDidMount(prevProps, prevState) {
    this.setTurnIndicatorFavicon(prevProps, prevState);
    this.refresh();
  }

  public componentWillUnmount() {
    document.getElementById('favicon').setAttribute('href', defaultFavicon);
    this.setState({ mounted: false });
  }

  public componentDidUpdate(prevProps, prevState) {
    this.setTurnIndicatorFavicon(prevProps, prevState);
  }

  private setTurnIndicatorFavicon(prevProps, prevState) {
    if (
      prevState?.game?.winning_team !== this.state.game?.winning_team ||
      prevState?.game?.round !== this.state.game?.round ||
      prevState?.game?.state_id !== this.state.game?.state_id
    ) {
      if (this.state.game?.winning_team) {
        document.getElementById('favicon').setAttribute('href', defaultFavicon);
      } else {
        document
          .getElementById('favicon')
          .setAttribute(
            'href',
            this.currentTeam() === 'blue' ? blueTurnFavicon : redTurnFavicon
          );
      }
    }
  }

  public refresh() {
    if (!this.state.mounted) {
      return;
    }

    let state_id = '';
    if (this.state.game && this.state.game.state_id) {
      state_id = this.state.game.state_id;
    }

    axios
      .post('/game-state', {
        game_id: this.props.gameID,
        state_id: state_id,
      })
      .then(({ data }) => {
        this.setState((oldState) => {
          const stateToUpdate = { game: data };
          return stateToUpdate;
        });
      })
      .finally(() => {
        setTimeout(() => {
          this.refresh();
        }, 2000);
      });
  }

  public nextWord(e, correct) {
    e.preventDefault();
    if (this.state.game.winning_team) {
      return; // ignore if game is over
    }
    axios
      .post('/next-word', {
        game_id: this.state.game.id,
        correct,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public currentPlayer() {
    return this.state.game.routingOrder[this.state.game.currentPlayer];
  }

  public currentTeam() {
    if (this.state.game.round % 2 == 0) {
      return this.state.game.starting_team;
    }
    return this.state.game.starting_team == 'red' ? 'blue' : 'red';
  }

  public remaining() {
    var count = 0;
    for (var i = 0; i < this.state.game.revealed.length; i++) {
      if (this.state.game.revealed[i]) {
        count++;
      }
    }
    return count;
  }

  public endTurn() {
    axios
      .post('/end-turn', {
        game_id: this.state.game.id,
        current_round: this.state.game.round,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public nextGame(e) {
    e.preventDefault();
    // Ask for confirmation when current game hasn't finished
    let allowNextGame =
      this.state.game.winning_team ||
      confirm('Do you really want to start a new game?');
    if (!allowNextGame) {
      return;
    }

    // TODO: Handle next game case => adding players automatically
    axios
      .post('/next-game', {
        game_id: this.state.game.id,
        word_set: this.state.game.word_set,
        create_new: true,
        timer_duration_ms: this.state.game.timer_duration_ms,
        enforce_timer: this.state.game.enforce_timer,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public handleAddWord(e, word) {
    e.preventDefault();

    axios
      .post('/add-word', {
        game_id: this.state.game.id,
        word: word,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public handleRemoveWord(e, word) {
    e.preventDefault();

    axios
      .post('/delete-word', {
        game_id: this.state.game.id,
        word: word,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public handleChangePlayerTeam(e, { name, team }) {
    e.preventDefault();

    axios
      .post('/change-player', {
        game_id: this.state.game.id,
        player_name: name,
        team: team,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public handleRemovePlayer(e, name) {
    e.preventDefault();

    axios
      .post('/delete-player', {
        game_id: this.state.game.id,
        player_name: name,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public handleNextStage(e) {
    e.preventDefault();

    axios
      .post('/start-game', {
        game_id: this.state.game.id,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  public handleAddPlayer(e, name) {
    e.preventDefault();

    axios
      .post('/add-player', {
        game_id: this.state.game.id,
        player_name: name,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      });
  }

  render() {
    const interstitialStages = [1, 3, 5];

    if (!this.state.game) {
      return <p className="loading">Loading&hellip;</p>;
    }

    let status, statusClass;
    if (this.state.game.winning_team) {
      statusClass = this.state.game.winning_team + ' win';
      status = this.state.game.winning_team + ' wins!';
    } else {
      statusClass = this.currentTeam() + '-turn';
      status = this.currentTeam() + "'s turn";
    }

    let endTurnButton;
    if (!this.state.game.winning_team) {
      endTurnButton = (
        <div id="end-turn-cont">
          <button
            onClick={(e) => this.endTurn()}
            id="end-turn-btn"
            aria-label={'End ' + this.currentTeam() + "'s turn"}
          >
            End {this.currentTeam()}&#39;s turn
          </button>
        </div>
      );
    }

    let otherTeam = 'blue';
    if (this.state.game.starting_team == 'blue') {
      otherTeam = 'red';
    }

    let shareLink = (
      <div id="share">
        Send this link to friends:&nbsp;
        <a className="url" href={window.location.href}>
          {window.location.href}
        </a>
      </div>
    );

    const timer = this.state.game.stage !== 0 && (
      <div id="timer">
        <Timer
          roundStartedAt={this.state.game.round_started_at}
          timerDurationMs={this.state.game.timer_duration_ms}
          handleExpiration={() => {
            this.state.game.enforce_timer && this.endTurn();
          }}
          freezeTimer={
            !!this.state.game.winning_team ||
            interstitialStages.includes(this.state.game.stage) ||
            this.state.game.stage === 0
          }
        />
      </div>
    );

    return (
      <div id="game-view">
        <div id="infoContent">
          {shareLink}
          {timer}
        </div>
        {this.state.game.stage === 0 && (
          <GameSetup
            words={this.state.game.words}
            handleAddWord={(e, word) => this.handleAddWord(e, word)}
            handleRemoveWord={(e, word) => this.handleRemoveWord(e, word)}
            players={this.state.game.team_players}
            handleChangePlayerTeam={(e, player) =>
              this.handleChangePlayerTeam(e, player)
            }
            handleRemovePlayer={(e, name) => this.handleRemovePlayer(e, name)}
            moveToNextStage={(e) => this.handleNextStage(e)}
            handleAddPlayer={(e, name) => this.handleAddPlayer(e, name)}
          />
        )}
        {interstitialStages.includes(this.state.game.stage) && (
          <StageEnd
            moveToNextStage={(e) => this.handleNextStage(e)}
            currentStage={this.state.game.stage}
            scores={this.state.game.team_points}
          />
        )}

        {this.state.game.stage !== 0 && <GameTurn />}
        <form id="mode-toggle" role="radiogroup">
          <button onClick={(e) => this.nextGame(e)} id="next-game-btn">
            Next game
          </button>
        </form>
        <div id="coffee">
          <a href="https://www.buymeacoffee.com/derrickpersson" target="_blank">
            Buy the developer a coffee.
          </a>
        </div>
      </div>
    );
  }
}
