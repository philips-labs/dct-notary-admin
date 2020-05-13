import React, { useState } from 'react';
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import { Grommet, Main, Grid } from 'grommet';
import logo from './logo.svg';
import { customTheme } from './Theme';
import { TargetsPage } from './pages';
import { NavBar, ApplicationContext, Notification } from './components';
import { setTimeout } from 'timers';

function App() {
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');

  const displayError = (message: string, autoHide: boolean) => {
    setError(message);
    if (autoHide) {
      setTimeout(() => setError(''), 3000);
    }
  };
  const displayInfo = (message: string, autoHide: boolean) => {
    setInfo(message);
    if (autoHide) {
      setTimeout(() => setInfo(''), 3000);
    }
  };

  return (
    <Grommet theme={customTheme} full>
      <Router>
        <Grid
          fill
          rows={['auto']}
          columns={['auto', 'flex']}
          areas={[
            { name: 'sidebar', start: [0, 0], end: [0, 0] },
            { name: 'main', start: [1, 0], end: [1, 0] },
          ]}
        >
          <ApplicationContext.Provider value={{ displayInfo, displayError }}>
            <NavBar gridArea="sidebar" />
            <Main gridArea="main" overflow="auto" pad="medium">
              {error ? <Notification type="error" message={error} /> : null}
              {info ? <Notification type="info" message={info} /> : null}
              <Switch>
                <Route exact path="/" component={Home} />
                <Route path="/targets" component={TargetsPage} />
              </Switch>
            </Main>
          </ApplicationContext.Provider>
        </Grid>
      </Router>
    </Grommet>
  );
}

function Home() {
  return (
    <>
      <img src={logo} className="App-logo" alt="logo" />
      <p>
        Edit <code>src/App.tsx</code> and save to reload.
      </p>
      <a className="App-link" href="https://reactjs.org" target="_blank" rel="noopener noreferrer">
        Learn React
      </a>
    </>
  );
}

export default App;
