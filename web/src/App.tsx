import { useState } from 'react';
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
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
    <Router>
      <ApplicationContext.Provider value={{ displayInfo, displayError }}>
        <div className="flex flex-row">
          <NavBar />
          <main className="p-5">
            {error ? <Notification type="error" message={error} /> : null}
            {info ? <Notification type="info" message={info} /> : null}
            <Switch>
              <Route path="/" component={TargetsPage} />
            </Switch>
          </main>
        </div>
      </ApplicationContext.Provider>
    </Router>
  );
}

export default App;
