import React, { PropTypes as T } from 'react';
import { Router } from 'react-router';
import { Provider } from 'react-redux';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import { Style } from 'radium';

class App extends React.Component {
  static contextTypes = {
    router: T.object
  }

  static propTypes = {
    history: T.object.isRequired,
    routes: T.element.isRequired,
    routerKey: T.number,
    actions: T.object
  };

  get content() {
    const { history, routes, routerKey, store, actions } = this.props;
    let newProps = {
      actions,
      ...this.props
    };

    const createElement = (Component, props) => {
      return <Component {...newProps} {...props} />;
    };

    return (
      <Provider store={store}>
        <Router
          key={routerKey}
          routes={routes}
          createElement={createElement}
          history={history} />
      </Provider>
    );
  }

  get devTools () {
    if (__DEBUG__) {
      if (!window.devToolsExtension) {
        const DevTools = require('containers/devtools').default;
        return <DevTools />;
      }
    }
  }

  render () {
    return (
       <Provider store={this.props.store}>
        <MuiThemeProvider>
          <div style={{height: '100%'}}>
            <Style rules={{body: {margin: '0px', fontFamily: 'Roboto, sans-serif', height: '100%'}, html: {height: '100%'}, '#root': {height: '100%'}}}/>
            {this.content}
            {this.devTools}
           </div>
        </MuiThemeProvider>
        </Provider>
     );
  }
}

export default App;
