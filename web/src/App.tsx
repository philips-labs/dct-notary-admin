import { useState } from 'react';
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import { Grommet } from 'grommet';
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
        <ApplicationContext.Provider value={{ displayInfo, displayError }}>
          <div className="flex flex-row">
            <NavBar />
            <main className="p-5">
              {error ? <Notification type="error" message={error} /> : null}
              {info ? <Notification type="info" message={info} /> : null}
              <Switch>
                <Route exact path="/" component={Home} />
                <Route path="/targets" component={TargetsPage} />
              </Switch>
            </main>
          </div>
        </ApplicationContext.Provider>
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
