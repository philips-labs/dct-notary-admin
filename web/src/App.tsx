import React from 'react';
import {
    BrowserRouter as Router,
    Switch,
    Route,
    Link
} from 'react-router-dom';
import logo from './logo.svg';
import './App.css';

function App() {
  return (
      <Router>
        <div className="App">
          <nav>
            <ul>
              <li>
                <Link to="/">Home</Link>
              </li>
              <li>
                <Link to="/targets">Targets</Link>
              </li>
            </ul>
          </nav>

          <main className="App-content">
            <Switch>
              <Route path="/targets">
                <Targets />
              </Route>
              <Route path="/">
                <Home />
              </Route>
            </Switch>
          </main>
        </div>
      </Router>
  );
}

function Home() {
  return (
    <>
      <img src={logo} className="App-logo" alt="logo" />
      <p>
        Edit <code>src/App.tsx</code> and save to reload.
      </p>
      <a
        className="App-link"
        href="https://reactjs.org"
        target="_blank"
        rel="noopener noreferrer"
      >
        Learn React
      </a>
    </>
  );
}

function Targets() {
  return <h2>Targets</h2>;
}

export default App;
