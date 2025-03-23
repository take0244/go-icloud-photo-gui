import React, { useState, useCallback } from "react";
import { LoginICloud } from "@/wailsjs/go/infraui/App";
import { useAlert } from 'react-alert';

function useStorageState(key, initialValue) {
  const [state, _setState] = useState(() => {
    try {
      const storedValue = localStorage.getItem(key);
      return storedValue ? JSON.parse(storedValue) : initialValue;
    } catch (error) {
      console.error("Error reading localStorage key:", key, error);
      return initialValue;
    }
  });

  const setState = useCallback(
    (v) => {
      _setState(v);
      localStorage.setItem(key, JSON.stringify(v));
    },
    [_setState, key]
  );

  return [state, setState];
}

const inputStyle = {
  width: "100%",
  margin: "10px 0",
  borderRadius: "8px",
  border: "1px solid #444",
  backgroundColor: "#222",
  color: "white",
  fontSize: "16px",
  outline: "none",
};

const buttonStyle = {
  width: "100%",
  padding: "10px",
  backgroundColor: "#007bff",
  color: "white",
  fontWeight: "bold",
  borderRadius: "8px",
  border: "none",
  cursor: "pointer",
  transition: "background 0.3s",
};

const containerStyle = {
  display: "flex",
  flexDirection: "column",
  alignItems: "center",
  justifyContent: "center",
  minHeight: "100vh",
  backgroundColor: "#121212",
  color: "white",
  fontFamily: "Arial, sans-serif",
};

const cardStyle = {
  padding: "24px",
  backgroundColor: "#1e1e1e",
  borderRadius: "12px",
  boxShadow: "0 4px 8px rgba(0, 0, 0, 0.3)",
  width: "320px",
  textAlign: "center",
};

// ログイン画面
export const Login = ({ setPage }) => {
  const alert = useAlert();
  const [username, setUsername] = useStorageState("username", "");
  const [password, setPassword] = useStorageState("password", "");
  const passwordRef = React.useRef(null);

  const handleLogin = () => {
    LoginICloud(username, password)
      .then((response) => {
        const resp = JSON.parse(response)
        if (resp?.error) {
          alert.error('ユーザー名かパスワードが違いますわ');
          return;
        }

        if (resp?.Required2fa) {
          setPage("code2fa");
          return;
        }

        setPage("photos");
      })
      .catch((error) => console.error("ログインエラー:", error));
  };

  return (
    <div style={containerStyle}>
      <div style={cardStyle}>
        <h2>ログイン</h2>
        {username}
        <input
          type="text"
          placeholder="ユーザー名"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          style={inputStyle}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              passwordRef.current?.focus();
            }
          }}
        />
        <input
          type="password"
          placeholder="パスワード"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          style={inputStyle}
          ref={passwordRef}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              handleLogin();
            }
          }}
        />
        <button
          onClick={handleLogin}
          style={buttonStyle}
          onMouseOver={(e) => (e.target.style.backgroundColor = "#0056b3")}
          onMouseOut={(e) => (e.target.style.backgroundColor = "#007bff")}
        >
          ログイン
        </button>
      </div>
    </div>
  );
};
