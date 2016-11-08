/* globals __DEBUG__ */
import React from 'react';
import ReactDOM from 'react-dom';

import App from 'containers/app';

import {browserHistory} from 'react-router';
//import makeRoutes from './routes';

import injectTapEventPlugin from 'react-tap-event-plugin';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

const initialState = {};
import {configureStore} from './configureStore';
const {store, actions, history} = configureStore({initialState, historyType: browserHistory});

let render = (routerKey = null) => {
  const makeRoutes = require('./routes').default;
  const routes = makeRoutes(store);

  const mountNode = document.querySelector('#root');
  ReactDOM.render(
    <App history={history}
          store={store}
          actions={actions}
          routes={routes}
          routerKey={routerKey} />, mountNode);
};

if (__DEBUG__ && module.hot) {
  const renderApp = render;
  render = () => renderApp(Math.random());

  module.hot.accept('./routes', () => render());
}

render();
