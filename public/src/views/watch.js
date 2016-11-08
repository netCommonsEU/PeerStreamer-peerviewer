import React, {PropTypes as T} from 'react';
import Player from 'containers/player';
import Radium from 'radium';
import Paper from 'material-ui/Paper';
import StreamInfo from 'containers/streaminfo';

const styles = {
  content: {
    maxWidth: '720px',
    maxHeight: '480px',
    margin: 'auto'
  },
  playerPaper: {
    marginBottom: '1em'
  },
  player: {
    width: '100%',
    height: '100%'
  },
  background: {
    backgroundColor: '#B0BEC5',
    height: '100%',
    margin: '0px',
    padding: '2em'
  }
};

const Watch = ({params}) => {
  return (
    <div style={styles.background}>
      <div style={styles.content}>
        <Paper style={styles.playerPaper}>
          <Player streamId={params.streamId} style={styles.player}/>
        </Paper>
        <StreamInfo streamId={params.streamId}/>
      </div>
    </div>
  );
};

export default Radium(Watch);
