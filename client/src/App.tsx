import React, { useState } from "react";
import { BrowserRouter as Router, Routes, Route, Link } from "react-router-dom";
import { Profile, Login, PublicProfile } from "./Profile";
import AtpAgent from "@atproto/api";

const App = () => {
  const [username, setUsername] = useState<string | undefined>();
  const [agent, setAgent] = useState<AtpAgent | null>(null);

  const handleLogin = (username: string, agent: AtpAgent) => {
    setUsername(username);
    setAgent(agent);
  };

  return (
    <Router>
      <div>
        <h1>Guestbook</h1>
        {username && agent ? (
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
              <Route
                path={`/${username}`}
                element={<Profile username={username} agent={agent} />}
              />
              <Route path="/" element={<h2>Welcome, {username}!</h2>} />
            </Routes>
            from
          </div>
        ) : (
          <Login onLogin={handleLogin} />
        )}
        <Routes>
          <Route path="/:username" element={<PublicProfile />} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;
