const NODE_ENV = process.env.NODE_ENV || 'development';
const dotenv = require('dotenv');

const webpack = require('webpack');
const path = require('path');

const join = path.join;
const resolve = path.resolve;

const getConfig = require('hjs-webpack');

const isDev = NODE_ENV === 'development';
const isTest = NODE_ENV === 'test';

// devServer config
const devHost   = process.env.HOST || 'localhost';
const devPort   = process.env.PORT || 8080;

const setPublicPath = process.env.SET_PUBLIC_PATH !== 'false';
//const publicPath  = (isDev && setPublicPath) ? `//${devHost}:${devPort}/` : '/';
const publicPath  = '/';

const root = resolve(__dirname);
const src = join(root, 'src');
const modules = join(root, 'node_modules');
const dest = join(root, 'dist');

var config = getConfig({
  isDev: isDev,
  in: join(src, 'app.js'),
  out: dest,
  html: function (context) {
    return {
      'index.html': context.defaultTemplate({
        title: 'PeerViewer',
        publicPath,
        meta: {}
      })
    };
  }
});

// ENV variables
//const dotEnvVars = dotenv.config();
const environmentEnv = dotenv.config({
  path: join(root, 'config', `${NODE_ENV}.config.js`),
  silent: true
});
const envVariables = Object.assign({}, environmentEnv);

const defines =
  Object.keys(envVariables)
  .reduce((memo, key) => {
    const val = JSON.stringify(envVariables[key]);
    memo[`__${key.toUpperCase()}__`] = val;
    return memo;
  }, {
    __NODE_ENV__: JSON.stringify(NODE_ENV),
    __DEBUG__: isDev
  });

config.plugins = [
  new webpack.DefinePlugin(defines)
].concat(config.plugins);
// END ENV variables

// Roots
config.resolve.root = [src, modules];
config.resolve.alias = {
  containers: join(src, 'containers'),
  components: join(src, 'components'),
  reducers: join(src, 'reducers'),
  actions: join(src, 'actions'),
  utils: join(src, 'utils'),
  api: join(src, 'api'),
  util: join(src, 'util')
};
// end Roots

// Dev
if (isDev) {
  config.devServer.port = devPort;
  config.devServer.hostname = devHost;
}

// Testing
if (isTest) {
  config.externals = {
    'react/addons': true,
    'react/lib/ReactContext': true,
    'react/lib/ExecutionEnvironment': true
  };

  config.module.noParse = /[/\\]sinon\.js/;
  config.resolve.alias.sinon = 'sinon/pkg/sinon';

  config.plugins = config.plugins.filter(p => {
    const name = p.constructor.toString();
    const fnName = name.match(/^function (.*)\((.*\))/);

    const idx = [
      'DedupePlugin',
      'UglifyJsPlugin'
    ].indexOf(fnName[1]);
    return idx < 0;
  });
}
// End Testing

module.exports = config;
