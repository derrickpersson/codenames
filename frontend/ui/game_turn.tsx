import * as React from 'react';
import { IPlayer, TeamPoint } from './models';

interface GameTurnProps {
  handleGetNextWord: (e, correct) => void;
  currentWord: string;
  currentStage: number;
  currentPlayer: IPlayer;
  scores: TeamPoint[];
  remaining: number;
  isYourTurn: boolean;
}

// Display current scores
// Display the current round
// Display whose turn it is
// Display the current word
// Display skip / correct commands
// Display how many words are left

const GameTurn: React.FunctionalComponent<GameTurnProps> = ({
  handleGetNextWord,
  currentWord,
  currentStage,
  currentPlayer,
  scores = [],
  remaining,
  isYourTurn = false,
}) => {
  const currentStageName = (current) => {
    switch (current) {
      case 2:
        return 'Explain';
      case 4:
        return 'Gestures';
      case 6:
        return 'One Word';
    }
  };

  const remainingCopy = (remaining) => {
    if (remaining >= 20) {
      return 'lots';
    } else if (remaining >= 10) {
      return 'some';
    } else {
      return `only ${remaining} left!`;
    }
  };

  return (
    <div>
      <div>
        <p>
          Current Stage: <strong>{currentStageName(currentStage)}</strong>
        </p>
        <p>Remaining: {remainingCopy(remaining)}</p>
      </div>
      <div style={{ margin: '2em 0' }}>
        {scores.map((score, idx) => (
          <div key={`${score.team}-${idx}`} style={{ margin: '1em 0' }}>
            {score.team} has <strong>{score.points}</strong> points
          </div>
        ))}
      </div>
      <div>
        Current Player:
        <div
          className={`tile ${currentPlayer.team}`}
          aria-label={currentPlayer.team}
          style={{ margin: '1em' }}
        >
          {isYourTurn ? 'You' : currentPlayer.player_name}
        </div>
      </div>
      {isYourTurn && (
        <div>
          Current Word:
          <div className={'tile'} style={{ fontSize: '1.5em' }}>
            {currentWord}
          </div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-around',
            }}
          >
            <button
              className="action-button pass"
              onClick={(e) => handleGetNextWord(e, false)}
            >
              Pass
            </button>
            <button
              className="action-button correct"
              onClick={(e) => handleGetNextWord(e, true)}
            >
              Correct
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default GameTurn;
