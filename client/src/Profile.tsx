import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { AtpAgent } from "@atproto/api";

const server = "http://localhost:9987";

// Mock API endpoints
const login = async (username: string, password: string) => {
  const agent = new AtpAgent({
    service: "https://bsky.social",
  });
  await agent.login({
    identifier: username,
    password: password,
  });
  return agent;
};

const getProfileData = async (did: string) => {
  try {
    const response = await fetch(`${server}/getProfileData/${did}`);
    if (!response.ok) throw new Error("Failed to fetch profile data");
    const data = await response.json();

    console.log("got response:", data);
    return data;
  } catch (error) {
    console.error("Error fetching comments:", error);
  }
};

const updateProfileData = (agent: AtpAgent, profileData: any) =>
  Promise.resolve();

const fetchComments = async (username: string) => {
  try {
    const response = await fetch(
      `${server}/getCommentsForPage/${username}`
    ).then((res) => res.json());
    return response;
  } catch (error) {
    console.error("Error fetching comments:", error);
    return [];
  }
};

const createComment = (token: string, commentText: string) => Promise.resolve();

// Login component
const Login = ({
  onLogin,
}: {
  onLogin: (username: string, agent: AtpAgent) => void;
}) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const agent = await login(username, password);
    onLogin(agent.session!.handle, agent);
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
const Profile = ({
  agent,
  username,
}: {
  agent: AtpAgent;
  username: string;
}) => {
  const [profile, setProfile] = useState({ text: "", links: [] });
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState("");

  useEffect(() => {
    getProfileData(username).then(setProfile);
    fetchComments(username).then(setComments);
  }, [agent, username]);

  const handleProfileSubmit = () => {
    updateProfileData(agent, profile).then(() => setIsEditing(false));
  };

  const handleCommentSubmit = () => {
    createComment(agent, newComment).then((createdComment) => {
      setComments([...comments, createdComment]);
      setNewComment("");
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
  const [loading, setLoading] = useState(true);
  const [profile, setProfile] = useState<{
    handle: "";
    text: "";
    links: [];
  } | null>(null);
  const [comments, setComments] = useState<
    {
      id: string;
      author: string;
      created: string;
      text: string;
    }[]
  >([]);

  useEffect(() => {
    if (!username) return;
    setLoading(true);
    fetchComments(username).then(setComments);

    getProfileData(username).then(setProfile);
    fetchComments(username);
    setLoading(false);
  }, [username]);

  if (!username) {
    return <div>No username provided</div>;
  }

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!profile) {
    return <div>Profile not found</div>;
  }

  return (
    <div>
      <h2>{profile.handle}'s LinkPage</h2>
      <p>{profile.text}</p>
      <ul>
        {profile.links.map((link) => (
          <li>
            <a href={link}>{link}</a>
          </li>
        ))}
      </ul>
      <h3>Comments</h3>
      <ul>
        {comments.map((comment) => (
          <li key={comment.id}>
            <p>
              From: {comment.author} at {comment.created}
            </p>
            <p>{comment.text}</p>
          </li>
        ))}
      </ul>
    </div>
  );
};

export { Profile, Login, PublicProfile };
