import OpenReplay from "@openreplay/tracker";

const projectKey = import.meta.env.VITE_OPENREPLAY_PROJECT_KEY || "";
const ingestPoint = import.meta.env.VITE_OPENREPLAY_INGEST_POINT || "https://api.openreplay.com/ingest";
const firstPartyAPIOrigin = import.meta.env.VITE_FIRST_PARTY_API_ORIGIN || window.location.origin;
const allowedOrigins = new Set([window.location.origin, firstPartyAPIOrigin]);

const tracker = projectKey
  ? new OpenReplay({
      projectKey,
      ingestPoint,
      revID: "demo-v1",
      respectDoNotTrack: true,
      obscureTextEmails: true,
      obscureTextNumbers: true,
      obscureInputEmails: true,
      obscureInputNumbers: true,
      obscureInputDates: true,
      defaultInputMode: 2,
      consoleMethods: ["warn", "error", "assert"],
      network: {
        failuresOnly: false,
        sessionTokenHeader: false,
        capturePayload: false,
        captureInIframes: false,
        ignoreHeaders: ["Authorization", "Cookie", "Set-Cookie", "X-API-Key"],
      },
      urls: {
        urlSanitizer: (value) => {
          const parsed = new URL(value, window.location.origin);
          return parsed.origin + parsed.pathname;
        },
      },
      __DISABLE_SECURE_MODE: import.meta.env.DEV,
    })
  : null;

let startPromise: Promise<void> | null = null;

export function startReplay(): Promise<void> {
  if (!tracker) return Promise.resolve();
  if (!startPromise) {
    startPromise = tracker.start({
      userID: "demo-user-918",
      metadata: {
        environment: "local",
        release: "demo-v1",
      },
    }).then(() => undefined).catch(() => undefined);
  }
  return startPromise;
}

export async function replayFetch(input: string, init?: RequestInit): Promise<Response> {
  await Promise.race([
    startReplay(),
    new Promise<void>((resolve) => window.setTimeout(resolve, 500)),
  ]);
  const target = new URL(input, window.location.origin);
  const headers = new Headers(init?.headers);
  if (tracker && allowedOrigins.has(target.origin)) {
    const sessionID = tracker.getSessionID();
    const sessionURL = tracker.getSessionURL({ withCurrentTime: true });
    if (sessionID) headers.set("X-OpenReplay-Session-ID", sessionID);
    if (sessionURL) headers.set("X-OpenReplay-Session-URL", sessionURL);
    headers.set("X-OpenReplay-Project-Key", projectKey);
  }
  return fetch(input, { ...init, headers });
}
