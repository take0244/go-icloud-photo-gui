import { useState, useEffect } from "react";
import { SelectDirectory, AllDownloadPhotos, Cancel } from "@/wailsjs/go/infraui/App";
import { useAlert } from 'react-alert';

export const Photos = ({ setPage }) => {
  const [selectedDir, setSelectedDir] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [progress, setProgress] = useState(null);
  const [phase, setPhase] = useState("");
  const alert = useAlert();

  useEffect(() => {
    window.runtime.EventsOn("app_progressEvent", (value) => {
      const progress = JSON.parse(value);
      setProgress(Math.floor(progress.value * 10000) / 100);
      setPhase(progress.phase);
    });
  }, []);

  const selectDirectory = async () => {
    const dir = await SelectDirectory();
    if (dir) {
      setSelectedDir(dir);
    }
  };

  const download = async () => {
    if (!selectedDir) return;
    setProgress(0);
    setIsLoading(true);
    try {
      const errorMessage = await AllDownloadPhotos(selectedDir);
      if (errorMessage) {
        alert.error(errorMessage);
        return;
      }
      alert.info('完了');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div style={{
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      minHeight: "100vh",
      backgroundColor: "#121212",
      color: "white",
      fontFamily: "Arial, sans-serif"
    }}>
      
      <div style={{
        padding: "24px",
        backgroundColor: "#1e1e1e",
        borderRadius: "12px",
        boxShadow: "0 4px 8px rgba(0, 0, 0, 0.3)",
        width: "320px",
        textAlign: "center"
      }}>
        <h2 style={{ fontSize: "20px", marginBottom: "16px" }}>写真をダウンロード</h2>

        <button
          style={{
            width: "100%",
            color: "white",
            fontWeight: "bold",
            padding: "10px",
            borderRadius: "8px",
            border: "none",
            transition: "background 0.3s",
            backgroundColor: isLoading ? "#555" : "#007bff",
            cursor: isLoading ? "not-allowed" : "pointer",
            opacity: isLoading ? 0.2 : 1,
          }}
          onClick={selectDirectory}
          disabled={isLoading}
          onMouseOver={(e) => e.target.style.backgroundColor = "#0056b3"}
          onMouseOut={(e) => e.target.style.backgroundColor = "#007bff"}
        >
          ダウンロード先を選択
        </button>

        {selectedDir && (
          <p style={{
            marginTop: "12px",
            fontSize: "14px",
            color: "#ccc",
            wordBreak: "break-all"
          }}>
            {selectedDir}
          </p>
        )}
        {phase === 'CHECK_FILES' && (
          <div>
            <p>ICloudのファイルを確認しています...</p>
            <p>ファイル数: {progress/100}</p>
          </div>
        )}
        {['DOWNLOAD_DUPLICATE', 'DOWNLOAD_ZIP'].includes(phase) && (
          <>
            <div style={{
              marginTop: "12px",
              width: "100%",
              backgroundColor: "#333",
              borderRadius: "8px",
              overflow: "hidden",
              marginBottom: "16px",
              height: "20px",
              position: "relative"
            }}>
              <div style={{
                width: `${progress}%`,
                backgroundColor: "#4caf50",
                height: "100%",
                transition: "width 0.3s ease-in-out"
              }} />
              <span style={{
                position: "absolute",
                width: "100%",
                textAlign: "center",
                fontSize: "12px",
                color: "#fff",
                fontWeight: "bold"
              }}>
              </span>
            </div>
            <p>phase: {phase}</p>
            {progress}%
          </>
        )}


        <button
          style={{
            width: "100%",
            marginTop: "16px",
            backgroundColor: isLoading || !selectedDir ? "#555" : "#28a745",
            color: "white",
            fontWeight: "bold",
            padding: "10px",
            borderRadius: "8px",
            border: "none",
            cursor: isLoading || !selectedDir ? "not-allowed" : "pointer",
            transition: "background 0.3s",
            opacity: isLoading || !selectedDir ? 0.2 : 1,
          }}
          onClick={download}
          disabled={!selectedDir || isLoading}
          onMouseOver={(e) => {
            if (!isLoading && selectedDir) e.target.style.backgroundColor = "#218838";
          }}
          onMouseOut={(e) => {
            if (!isLoading && selectedDir) e.target.style.backgroundColor = "#28a745";
          }}
        >
          {isLoading ? "ダウンロード中..." : "ダウンロード開始"}
        </button>

        <button
          style={{
            width: "100%",
            marginTop: "16px",
            backgroundColor: "#F00",
            color: "white",
            fontWeight: "bold",
            padding: "10px",
            borderRadius: "8px",
            border: "none",
            cursor: "pointer",
            transition: "background 0.3s",
          }}
          onClick={() => {
            Cancel().then(() => {
              setPage("login");
            })
          }}
        >
          終了
        </button>
      </div>
    </div>
  );
};
