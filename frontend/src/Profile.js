import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useParams } from 'react-router-dom';
import axios from 'axios';
import { BskyAgent } from '@atproto/api'

const server = 'http://localhost:9987'

// Mock API endpoints
const login = async (username, password) => {
	const agent = new BskyAgent({
		service: "https://bsky.social",
	});
	await agent.login({
		identifier: username,
		password: password,
	})
	return agent
}


const getProfileData = async (did) => {
	try {
		const response = await axios.get(`${server}/getProfileData/${did}`);
		console.log("got response: ", response.data)
		return response.data
	} catch (error) {
		console.error('Error fetching comments:', error);
	}
}

const updateProfileData = (agent, profileData) => Promise.resolve();

const fetchComments = async (username) => {
	try {
		const response = await axios.get(`${server}/getCommentsForPage/${username}`);
		return response.data;
	} catch (error) {
		console.error('Error fetching comments:', error);
		return [];
	}
};

const createComment = (token, commentText) => Promise.resolve();

// Login component
const Login = ({ onLogin }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [agent, setAgent] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    let la = await login(username, password);
    setAgent(la)
    console.log(la);
    onLogin(la.session.handle);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      <button type="submit">Log in</button>
    </form>
  );
};

// Profile component
const Profile = ({ agent, username }) => {
  const [profile, setProfile] = useState({ text: '', links: [] });
  const [isEditing, setIsEditing] = useState(false);
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');

  useEffect(() => {
    getProfileData(agent).then(setProfile);
    fetchComments(username).then(setComments);
  }, [agent]);

  const handleProfileSubmit = () => {
    updateProfileData(agent, profile).then(() => setIsEditing(false));
  };

  const handleCommentSubmit = () => {
    createComment(agent, newComment).then((createdComment) => {
      setComments([...comments, createdComment]);
      setNewComment('');
    });
  };

  return (
    <div>
      <h2>Welcome, {username}!</h2>
      {/* Rest of the Profile component code */}
    </div>
  );
};

const PublicProfile = () => {
  const { username } = useParams();
  const [profile, setProfile] = useState({ text: '', links: [] });
  const [comments, setComments] = useState([]);

  useEffect(() => {
    fetchComments(username).then(setComments);

    getProfileData(username).then(setProfile);
    fetchComments();
  }, [username]);

  return (
    <div>
      <h2>{profile.handle}'s LinkPage</h2>
      <p>{profile.text}</p>
      <ul>
        {profile.links.map((link) => (
          <li >
            <a href={link}>{link}</a>
          </li>
        ))}
      </ul>
      <h3>Comments</h3>
	  <ul>
      {comments.map((comment) => (
	      <li key={comment.id}>
	<p>From: {comment.author} at {comment.created}</p>
        <p>{comment.text}</p>
	      </li>
      ))}
	  </ul>
    </div>
  );
};


export {
	Profile,
	Login,
	PublicProfile,
}
