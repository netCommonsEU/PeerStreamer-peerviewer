import { combineReducers } from 'redux';

const byId = (state = {}, action) => {
  switch (action.type) {
  case 'RECEIVE_STREAMS':
    return action.response.entities.streams;
  default:
    return state;
  }
};

const ids = (state = [], action) => {
  switch (action.type) {
  case 'RECEIVE_STREAMS':
    return action.response.result;
  default:
    return state;
  }
};

const loaded = (state = false, action) => {
  switch (action.type) {
  case 'RECEIVE_STREAMS':
    return true;
  default:
    return state;
  }
};

export default combineReducers({byId, ids, loaded});

export const getStreamsLoaded = (state) => {
  return state.loaded;
};

export const getAvailableStreams = (state) => {
  return state.ids.map(id => {
    return state.byId[id];
  });
};

export const getStreamDetails = (state, id) => {
  return state.byId[id];
};
