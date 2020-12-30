import * as React from 'react';
import axios from 'axios';
import WordSetToggle from '~/ui/wordset_toggle';
import TimerSettings from '~/ui/timer_settings';
import OriginalWords from '~/words.json';
import Toggle from '~/ui/toggle';

export const Lobby = ({ defaultGameID }) => {
  const [playerName, setPlayerName] = React.useState('');
  const [newGameName, setNewGameName] = React.useState(defaultGameID);
  const [enableRandomWords, setEnableRandomWords] = React.useState(false);
  const [selectedWordSets, setSelectedWordSets] = React.useState(['English']);
  const [customWordsText, setCustomWordsText] = React.useState('');
  const [words, setWords] = React.useState({ ...OriginalWords, Custom: [] });
  const [warning, setWarning] = React.useState(null);
  const [timer, setTimer] = React.useState([1, 0]);
  const [enforceTimerEnabled, setEnforceTimerEnabled] = React.useState(false);

  let selectedWordCount = selectedWordSets
    .map((l) => words[l].length)
    .reduce((a, cv) => a + cv, 0);

  React.useEffect(() => {
    if (selectedWordCount >= 25) {
      setWarning(null);
    }
  }, [selectedWordSets, customWordsText]);

  function handleNewGame(e) {
    e.preventDefault();
    if (!newGameName || !playerName) {
      return;
    }

    let combinedWordSet = selectedWordSets
      .map((l) => words[l])
      .reduce((a, w) => a.concat(w), []);

    if (combinedWordSet.length < 25) {
      setWarning('Selected wordsets do not include at least 25 words.');
      return;
    }

    axios
      .post('/next-game', {
        game_id: newGameName,
        word_set: enableRandomWords ? combinedWordSet : [],
        create_new: false,
        timer_duration_ms:
          timer && timer.length ? timer[0] * 60 * 1000 + timer[1] * 1000 : 0,
        enforce_timer: true,
        player_name: playerName,
      })
      .then(() => {
        const newURL = (document.location.pathname = '/' + newGameName);
        window.location = newURL;
      });
  }

  let toggleWordSet = (wordSet) => {
    let wordSets = [...selectedWordSets];
    let index = wordSets.indexOf(wordSet);

    if (index == -1) {
      wordSets.push(wordSet);
    } else {
      wordSets.splice(index, 1);
    }
    setSelectedWordSets(wordSets);
  };

  let langs = Object.keys(OriginalWords);
  langs.sort();

  return (
    <div id="lobby">
      <div id="available-games">
        <form id="new-game">
          <p className="intro">
            Play bowls online with friends. To create a new game or join an
            existing game, enter your name and a game identifier and click 'GO'.
          </p>
          <input
            type="text"
            id="player-name"
            aria-label="player name"
            autoFocus
            onChange={(e) => {
              setPlayerName(e.target.value);
            }}
            value={playerName}
            placeholder={'Your name'}
          />
          <input
            type="text"
            id="game-name"
            aria-label="game identifier"
            autoFocus
            onChange={(e) => {
              setNewGameName(e.target.value);
            }}
            value={newGameName}
          />

          <button
            disabled={!newGameName.length || !playerName.length}
            onClick={handleNewGame}
          >
            Go
          </button>

          {warning !== null ? (
            <div className="warning">{warning}</div>
          ) : (
            <div></div>
          )}
          <TimerSettings
            {...{
              timer,
              setTimer,
              enforceTimerEnabled,
              setEnforceTimerEnabled,
            }}
          />

          <div className="toggle-container">
            <span>Use Random Words:</span>
            <Toggle
              name="enable random words"
              state={enableRandomWords}
              handleToggle={() => setEnableRandomWords(!enableRandomWords)}
            />
          </div>

          {enableRandomWords && (
            <div id="new-game-options">
              <div id="wordsets">
                <p className="instruction">
                  You've selected <strong>{selectedWordCount}</strong> words.
                </p>
                <div id="default-wordsets">
                  {langs.map((_label) => (
                    <WordSetToggle
                      key={_label}
                      words={words[_label]}
                      label={_label}
                      selected={selectedWordSets.includes(_label)}
                      onToggle={(e) => toggleWordSet(_label)}
                    ></WordSetToggle>
                  ))}
                </div>
              </div>
            </div>
          )}
        </form>
      </div>
    </div>
  );
};
