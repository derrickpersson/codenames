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
  usesRandomWords: boolean;
}

const GameSetup: React.FunctionalComponent<GameSetupProps> = ({
  words = [],
  handleAddWord,
  handleRemoveWord,
  players = [],
  handleChangePlayerTeam,
  handleRemovePlayer,
  moveToNextStage,
  usesRandomWords,
}) => {
  const [privateWords, setPrivateWords] = React.useState([]);
  const [wordInput, setWordInput] = React.useState('');
  const teamComposition = players.reduce(
    (a, player) => {
      a[player.team]++;
      return a;
    },
    { red: 0, blue: 0 }
  );
  const hasOpposingTeams =
    teamComposition.red >= 2 && teamComposition.blue >= 2;
  return (
    <div>
      <div className="setup-container">
        <div className="column">
          <div>Custom Words: ({words.length} total)</div>
          <div>
            {privateWords.map((word, idx) => (
              <div key={`${word}-${idx}`} className="tile">
                {word}

                <button
                  className="remove"
                  onClick={(e) => {
                    handleRemoveWord(e, word);
                    setPrivateWords([
                      ...privateWords.slice(0, idx),
                      ...privateWords.slice(idx + 1),
                    ]);
                  }}
                >
                  +
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
                className={`tile ${player.team} playerTile`}
                key={`${player.player_name}-${idx}`}
              >
                {player.player_name}
                <div className="playerActionsContainer">
                  <button
                    className={`changeTeams ${player.team}Switch`}
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
                    +
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className="action-footer">
        <button
          disabled={!hasOpposingTeams || (words.length > 5 && !usesRandomWords)}
          onClick={(e) => moveToNextStage(e)}
        >
          Start
        </button>
      </div>
    </div>
  );
};

export default GameSetup;
