import * as streamsAPI from 'api/streams';
import { normalize } from 'normalizr';
import * as schema from './schema';

export const fetchStreams = () => (dispatch) => {
  streamsAPI.fetchStreams().then(response => {
    dispatch({
      type: 'RECEIVE_STREAMS',
      response: normalize(response, schema.arrayOfStreams)
    });
  });
};
