import { connect } from 'react-redux';
import StreamInfo from 'components/streaminfo';
import { getStreamDetails } from 'reducers';

const mapStateToProps = (state, {streamId}) => {
    const details = getStreamDetails(state, streamId);
    if (!details) {
        return {
            title: '[Not available]'
        }
    }
    return {
        title: details.description
    }
};

export default connect(mapStateToProps)(StreamInfo)