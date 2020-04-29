import React from 'react';
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import { Grommet, Main, Grid } from 'grommet';
import logo from './logo.svg';
import { customTheme } from './Theme';
import { TargetsPage } from './pages';
import { NavBar } from './components';

function App() {
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
          <NavBar gridArea="sidebar" />
          <Main gridArea="main" overflow="auto" pad="medium">
            <Switch>
              <Route exact path="/" component={Home} />
              <Route path="/targets" component={TargetsPage} />
            </Switch>
          </Main>
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
