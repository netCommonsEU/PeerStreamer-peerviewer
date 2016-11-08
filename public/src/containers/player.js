import { connect } from 'react-redux';
import { getStreamDetails, getStreamsLoaded } from 'reducers';
import { push } from 'react-router-redux';
import Player from 'components/player';
import { fetchStreams } from 'actions/streams';

const mapStateToProps = (state, {streamId}) => ({
  ...getStreamDetails(state, streamId),
  loaded: getStreamsLoaded(state)
});

export default connect(mapStateToProps, {goHome: () => (push('/')), fetchStreams})(Player);
