import * as React from 'react';
import { TeamPoint } from './models';

interface StageEndProps {
  moveToNextStage: (e) => void;
  currentStage: number;
  scores: TeamPoint[];
}

// Display current scores
// Display the current round
// Display whose turn it is next

const StageEnd: React.FunctionalComponent<StageEndProps> = ({
  moveToNextStage,
  currentStage,
  scores = [],
}) => {
  const nextStage = (current) => {
    switch (current) {
      case 1:
        return 'Explain';
      case 3:
        return 'Gestures';
      case 5:
        return 'One Word';
    }
  };

  return (
    <div className="container" style={{ margin: '2em 0' }}>
      <div>
        Next Stage: <strong>{nextStage(currentStage)}</strong>
      </div>
      <div>
        {scores.map((score, idx) => (
          <div key={`${score.team}-${idx}`}>
            {score.team} has {score.points}
          </div>
        ))}
      </div>
      <div
        style={{
          margin: '2em 0',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        Get Ready for the next stage!
      </div>
      <div className="action-footer">
        <button onClick={(e) => moveToNextStage(e)}>Start</button>
      </div>
    </div>
  );
};

export default StageEnd;
