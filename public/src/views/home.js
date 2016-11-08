import React, { PropTypes as T } from 'react';
import {Link} from 'react-router';
import AvailableStreamList from 'containers/availablestreamlist';
import Radium from 'radium';

const styles = {
  container: {
    margin: '2em',
    maxWidth: '1024px',
  }
}

export class IndexPage extends React.Component {
  render() {
    return (
      <div style={styles.container}>
        <h1>Available channels</h1>
        <p>Choose one channel and enjoy it directly in your browser!</p>
        <AvailableStreamList />
      </div>
    );
  }
}

export default Radium(IndexPage);
