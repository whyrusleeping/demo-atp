import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Profile, Login, PublicProfile } from './Profile';

const App = () => {
  const [username, setUsername] = useState(null);

  const handleLogin = (username) => {
    setUsername(username);
  };

  return (
    <Router>
      <div>
        <h1>Guestbook</h1>
        {username ? (
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
              <Route path={`/${username}`} element={<Profile username={username} />} />
              <Route path="/" element={ <h2>Welcome, {username}!</h2> } />
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
