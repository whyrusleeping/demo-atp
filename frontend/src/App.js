import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Profile, Login, PublicProfile } from './Profile';

const App = () => {
  const [token, setToken] = useState(null);
  const [username, setUsername] = useState(null);

  const handleLogin = (token, username) => {
    setToken(token);
    setUsername(username);
  };

  return (
    <Router>
      <div>
        <h1>Guestbook</h1>
        {token ? (
          <div>
            <nav>
              <ul>
                <li>
                  <Link to={`/${username}`}>My Profile</Link>
                </li>
                <li>
                  <Link to="/">Home</Link>
                </li>
              </ul>
            </nav>
            <Routes>
              <Route path={`/${username}`}>
                <Profile token={token} username={username} />
              </Route>
              <Route path="/">
                <h2>Welcome, {username}!</h2>
              </Route>
            </Routes>
          </div>
        ) : (
          <Login onLogin={handleLogin} />
        )}
        <Routes>
          <Route path="/:username" element={ <PublicProfile /> } />
        </Routes>
      </div>
    </Router>
  );
};

export default App;
