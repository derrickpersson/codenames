import * as React from 'react';
import { IPlayer, Team } from './models';
import '~/game_setup.css';

interface GameSetupProps {
  words: string[];
  handleAddWord: (e, word: string) => void;
  handleRemoveWord: (e, word: string) => void;
  players: IPlayer[];
  handleChangePlayerTeam: (e, { name: string, team: Team }) => void;
  handleRemovePlayer: (e, name) => void;
  handleAddPlayer: (e, name) => void;
  moveToNextStage: (e) => void;
}

// What happens on this screen?
// Display teams (members on each)
// Can change your own team
// Add words

const GameSetup: React.FunctionalComponent<GameSetupProps> = ({
  words = [],
  handleAddWord,
  handleRemoveWord,
  players = [],
  handleChangePlayerTeam,
  handleRemovePlayer,
  handleAddPlayer,
  moveToNextStage,
}) => {
  const [privateWords, setPrivateWords] = React.useState([]);
  const [wordInput, setWordInput] = React.useState('');
  const [playerInput, setPlayerInput] = React.useState('');
  const teamComposition = players.reduce(
    (a, player) => {
      a[player.team] = true;
      return a;
    },
    { red: false, blue: false }
  );
  const hasOpposingTeams = teamComposition.red && teamComposition.blue;

  return (
    <div>
      <div className="setup-container">
        <div className="column">
          <div>Words: ({words.length} total)</div>
          <div>
            {privateWords.map((word, idx) => (
              <div key={`${word}-${idx}`} className="tile">
                {word}
                <button
                  className="remove"
                  onClick={(e) => {
                    handleRemoveWord(e, word);
                  }}
                >
                  X
                </button>
              </div>
            ))}
          </div>
          <div className="addContainer">
            <input
              className="addInput"
              value={wordInput}
              onChange={(e) => setWordInput(e.target?.value)}
            />
            <button
              className="add"
              disabled={wordInput.length === 0}
              onClick={(e) => {
                handleAddWord(e, wordInput);
                setWordInput('');
                setPrivateWords([...privateWords, wordInput]);
              }}
            >
              Add Phrase
            </button>
          </div>
        </div>
        <div className="column">
          <div>
            <div>Players:</div>
            {(players || []).map((player, idx) => (
              <div
                className={`tile ${player.team}`}
                key={`${player.player_name}-${idx}`}
              >
                {player.player_name}
                <div>
                  <button
                    onClick={(e) =>
                      handleChangePlayerTeam(e, {
                        name: player.player_name,
                        team: player.team === 'blue' ? 'red' : 'blue',
                      })
                    }
                  >
                    Change Team
                  </button>
                  <button
                    className="remove"
                    onClick={(e) => {
                      handleRemovePlayer(e, player.player_name);
                    }}
                  >
                    X
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className="action-footer">
        <button
          disabled={words.length < 5 || !hasOpposingTeams}
          onClick={(e) => moveToNextStage(e)}
        >
          Start
        </button>
      </div>
    </div>
  );
};

export default GameSetup;
