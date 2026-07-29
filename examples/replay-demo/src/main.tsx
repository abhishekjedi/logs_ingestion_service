import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { startReplay } from "./openreplay";
import "./style.css";

void startReplay();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
