import React from 'react';
import {Route, IndexRoute} from 'react-router';
import Home from 'views/home';
import Navigation from 'containers/navigation';

export const makeRoutes = () => {
  return (
    <Route path='/' component={Navigation}>
      {/* Lazy-loading */}
      <Route path="watch/:streamId" getComponent={(location, cb) => {
        require.ensure([], (require) => {
          const mod = require('./views/watch');
          cb(null, mod.default);
        });
      }} />
      <Route path="about" getComponent={(location, cb) => {
        require.ensure([], (require) => {
          const mod = require('views/about');
          cb(null, mod.default);
        });
      }} />
      {/* inline loading */}
      <IndexRoute component={Home} />
    </Route>
  );
};

export default makeRoutes;
