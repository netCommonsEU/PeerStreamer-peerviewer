/* globals __DEBUG__ */
import { browserHistory } from 'react-router';
import { routerMiddleware, syncHistoryWithStore } from 'react-router-redux';
import thunkMiddleware from 'redux-thunk';
import { createStore, compose, applyMiddleware } from 'redux';
import rootReducer from 'reducers';

export const configureStore = ({
  historyType = browserHistory,
  userInitialState = {}}) => {

  let middleware = [
    thunkMiddleware,
    routerMiddleware(historyType)
  ];

  let tools = [];
  if (__DEBUG__) {
    const DevTools = require('containers/devtools').default;
    let devTools = window.devToolsExtension ? window.devToolsExtension : DevTools.instrument;
    if (typeof devTools === 'function') {
      tools.push(devTools());
    }
  }

  let finalCreateStore;
  finalCreateStore = compose(
      applyMiddleware(...middleware),
      ...tools
    )(createStore);

  const store = finalCreateStore(
      rootReducer,
      Object.assign({}, userInitialState)
    );

  const history = syncHistoryWithStore(historyType, store, {
    adjustUrlOnReplay: true
  });

  if (module.hot) {
    module.hot.accept('reducers', () => {
      const {rootReducer} = require('reducers');
      store.replaceReducer(rootReducer);
    });
  }

  return {store, history};
};
