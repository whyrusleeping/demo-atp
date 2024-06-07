import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useParams } from 'react-router-dom';
import axios from 'axios';

// Mock API endpoints
const login = (username, password) => Promise.resolve({
  token: 'example-token',
  username: username,
});

const getProfileData = async (did) => {
	try {
		const response = await axios.get(`http://localhost:9987/getProfileData/${did}`);
		return response.data
	} catch (error) {
		console.error('Error fetching comments:', error);
	}
}

const updateProfileData = (token, profileData) => Promise.resolve();

const getCommentsForPage = (token, page) => Promise.resolve([
  { id: 1, text: 'Great profile!' },
  { id: 2, text: 'Thanks for sharing!' },
]);

const createComment = (token, commentText) => Promise.resolve();

// Login component
const Login = ({ onLogin }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    login(username, password).then(({ token, username }) => {
      onLogin(token, username);
    });
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
const Profile = ({ token, username }) => {
  const [profile, setProfile] = useState({ text: '', links: [] });
  const [isEditing, setIsEditing] = useState(false);
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');

  useEffect(() => {
    getProfileData(token).then(setProfile);
    getCommentsForPage(token, 1).then(setComments);
  }, [token]);

  const handleProfileSubmit = () => {
    updateProfileData(token, profile).then(() => setIsEditing(false));
  };

  const handleCommentSubmit = () => {
    createComment(token, newComment).then((createdComment) => {
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
    // Fetch profile data for the specified username
    const fetchComments = async () => {
      try {
        const response = await axios.get(`http://localhost:9987/getCommentsForPage/${username}`);
        setComments(response.data);
      } catch (error) {
        console.error('Error fetching comments:', error);
      }
    };

    getProfileData(username).then(setProfile);
    fetchComments();
  }, [username]);

  return (
    <div>
      <h2>{profile.handle}'s Profile</h2>
      <p>{profile.text}</p>
      <ul>
        {profile.links.map((link) => (
          <li >
            <a href={link}>{link}</a>
          </li>
        ))}
      </ul>
      <h3>Comments</h3>
      {comments.map((comment) => (
        <p key={comment.id}>{comment.text}</p>
      ))}
    </div>
  );
};


export {
	Profile,
	Login,
	PublicProfile,
}
