import { connect } from 'react-redux';
import { getAvailableStreams } from 'reducers';
import { fetchStreams } from 'actions/streams';
import StreamList from 'components/streamlist';
import { push } from 'react-router-redux';

const mapStateToProps = (state) => {
  return {
    streams: getAvailableStreams(state)
  };
};

export default connect(mapStateToProps, {fetchStreams, navigateTo: push})(StreamList);
