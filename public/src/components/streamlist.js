import React, {PropTypes as T} from 'react';
import { Link } from 'react-router';
import {List, ListItem} from 'material-ui/List';
import Subheader from 'material-ui/Subheader';
import Divider from 'material-ui/Divider';
import LiveTv from 'material-ui/svg-icons/notification/live-tv';
import Radio from 'material-ui/svg-icons/av/radio';

const navigateGenerator = (push, streamId) => () => push(`/watch/${streamId}`);

export default class StreamList extends React.Component {
  componentDidMount() {
    this.props.fetchStreams();
  }

  render() {
    const { streams, navigateTo } = this.props;
    const videoChannels = streams.filter(e => e.mediaType == 'video');
    const audioChannels = streams.filter(e => e.mediaType == 'audio');
    const audioChannelItems = audioChannels.map(stream => (
      <ListItem key={stream.id} primaryText={stream.description} leftIcon={<Radio />} onTouchTap={navigateGenerator(navigateTo, stream.id)}/>
    ));
    const videoChannelItems = videoChannels.map(stream => (
      <ListItem key={stream.id} primaryText={stream.description} leftIcon={<LiveTv />} onTouchTap={navigateGenerator(navigateTo, stream.id)}/>
    ));
    return (
      <div>
        <List>
          <Subheader>Video channels</Subheader>
          {videoChannelItems}
        </List>
        <Divider />
        <List>
          <Subheader>Audio channels</Subheader>
          {audioChannelItems}
        </List>
      </div>
    );
  }
}
