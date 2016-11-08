import { connect } from 'react-redux';
import Navigation from 'components/navigation';
import { getStreamDetails } from 'reducers';
import { setDrawerState } from 'reducers/navigation';
import { push } from 'react-router-redux';

const getCurrentTitle = (state, props) => {
  const { location: { pathname }, params: { streamId }} = props;
  let title = 'PeerViewer';
  let documentTitle = 'PeerViewer';
  switch (false) {
    case !/^\/watch/.test(pathname):
      {
        const details = getStreamDetails(state, streamId);
        if (!details) {
          title = 'Loading...';
          break;
        }
        switch (details.mediaType) {
          case 'video':
            title = 'Now watching';
            break;
          case 'audio':
            title = 'Now listening';
            break;
          default:
            title = 'Stream';
        }
        documentTitle = `${details.description} - PeerViewer`;
      }
      break;
    case !/^\/about/.test(pathname):
      title = 'About';
      documentTitle = 'About - PeerViewer'
  }
  document.title = documentTitle;
  return title;
};

const mapStateToProps = (state, props) => {
  return {
    title: getCurrentTitle(state, props),
    drawerOpen: state.navigation.drawerOpen
  };
};

const mapDispatchToProps = {
  setDrawerState,
  navigateTo: push
};

export default connect(mapStateToProps, mapDispatchToProps)(Navigation);