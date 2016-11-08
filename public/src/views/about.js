import React from 'react';
import Radium from 'radium';

const styles = {
  container: {
    margin: '2em',
    maxWidth: '1023px'
  },

};

export const About = () => (
  <div style={styles.container}>
    <h1>PeerViewer</h1>
    <h2>A media player for distributed environments</h2>
    <p>I developed this application as a part of my thesis work. This project is based on pure Go implementations of the <i>GRAPES</i><sup>[1]</sup> and <i>Messaging Library</i><sup>[2]</sup>, two components originally written for the <i>PeerStreamer</i><sup>[3]</sup> project.</p>

    <p>All the source code is free and open. Please find it on GitHub:</p>
    <ul>
      <li><a href="https://github.com/netCommonsEU/PeerStreamer-peerviewer"><pre>peerviewer</pre></a>A media player for distributed environments</li>
      <li><a href="https://github.com/netCommonsEU/PeerStreamer-go-grapes"><pre>go-grapes</pre></a>A pure Go implementation of the GRAPES library</li>
      <li><a href="https://github.com/netCommonsEU/PeerStreamer-go-ml"><pre>go-ml</pre></a>A pure Go implementation of the Messaging Library</li>
    </ul>

    <h3>License</h3>
    This work is licensed under the terms of the GPLv3 license. <a href="https://www.gnu.org/licenses/gpl.html">Here</a>'s a full copy of the license text.

    <h3>Notes</h3>
    <ol>
      <li><a href="https://github.com/netCommonsEU/PeerStreamer-grapes">GRAPES homepage</a></li>
      <li><a href="https://github.com/netCommonsEU/PeerStreamer-napa-baselibs"> Messaging Library</a></li>
      <li><a href="https://github.com/netCommonsEU/PeerStreamer">PeerStreamer: fast and efficient P2P streaming</a></li>
    </ol>
  </div>
);

export default Radium(About);
