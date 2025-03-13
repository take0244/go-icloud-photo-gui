import { useState } from "react";
import { Code2fa } from "@/wailsjs/go/infraui/App";
import { useAlert } from 'react-alert';

const inputStyle = {
  padding: "10px",
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
  color: "white",
  fontWeight: "bold",
  borderRadius: "8px",
  border: "none",
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
export const Code2Fa = ({ setPage }) => {
  const [twoFactorCode, setTwoFactorCode] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const alert = useAlert();
  const handleVerifyTwoFactor = () => {
    setIsLoading(true)
    Code2fa(twoFactorCode)
      .then((res) => {
        if (JSON.parse(res)) {
          setPage("photos");
        } else {
          alert.show("認証コードが違います。やり直してね。")
          setPage("login");
        }
      })
      .catch(() => {
        alert.show("エラーが発生しました。ログインからやり直してください。")
        setPage("login");
      })
      .finally(() => {
        setIsLoading(false);
      });
  };

  return (
    <div style={containerStyle}>
      <div style={cardStyle}>
        <h3>2段階認証コードを入力</h3>
        <input
          type="text"
          placeholder="認証コード"
          value={twoFactorCode}
          onChange={(e) => setTwoFactorCode(e.target.value?.trim())}
          style={inputStyle}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              handleVerifyTwoFactor();
            }
          }}
        />
        <button
          onClick={handleVerifyTwoFactor}
          style={{...buttonStyle,
            backgroundColor: isLoading ? "#555" : "#007bff",
            cursor: isLoading ? "not-allowed" : "pointer",
            opacity: isLoading ? 0.2 : 1,
          }}
          disabled={isLoading}
          onMouseOver={(e) => (e.target.style.backgroundColor = "#0056b3")}
          onMouseOut={(e) => (e.target.style.backgroundColor = "#007bff")}
        >
          送信
        </button>
      </div>
    </div>
  );
};
