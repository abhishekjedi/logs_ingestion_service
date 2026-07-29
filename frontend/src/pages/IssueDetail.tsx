import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  api,
  type Issue,
  type TimePoint,
  type ErrorEvent,
  type Breadcrumb,
  type SessionContext,
  type SessionContextEvent,
} from "../api";
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";

const hour = (t: string) => new Date(t).toLocaleString([], { month: "short", day: "numeric", hour: "2-digit" });
const chartTip = { background: "#1e222d", border: "1px solid #272c3a", borderRadius: 8, color: "#e6e9ef" };

export default function IssueDetail() {
  const { issueId } = useParams();
  const id = Number(issueId);
  const nav = useNavigate();
  const [issue, setIssue] = useState<Issue | null>(null);
  const [points, setPoints] = useState<TimePoint[]>([]);
  const [events, setEvents] = useState<ErrorEvent[]>([]);

  useEffect(() => {
    api.getIssue(id).then((r) => setIssue(r.issue));
    api.issueTimeseries(id).then((r) => setPoints(r.points || []));
    api.issueEvents(id).then((r) => setEvents(r.events || []));
  }, [id]);

  if (!issue) return <div className="muted">Loading…</div>;
  const data = points.map((p) => ({ t: hour(p.timestamp), events: p.events }));

  return (
    <div className="stack">
      <div>
        <a className="muted" style={{ fontSize: 13, cursor: "pointer" }} onClick={() => nav(-1)}>
          ← Back
        </a>
        <div className="row" style={{ marginTop: 8, gap: 10 }}>
          <span className={`badge lvl-${issue.level}`}>{issue.level}</span>
          <span className={`badge st-${issue.status}`}>{issue.status}</span>
        </div>
        <h1 className="page-title" style={{ marginTop: 10 }}>
          {issue.title || "(untitled)"}
        </h1>
        {issue.culprit && <div className="muted">{issue.culprit}</div>}
      </div>

      <div className="grid cols3">
        <div className="stat">
          <div className="k">Events</div>
          <div className="v">{issue.event_count.toLocaleString()}</div>
        </div>
        <div className="stat">
          <div className="k">Affected users</div>
          <div className="v">{issue.affected_users_estimate.toLocaleString()}</div>
        </div>
        <div className="stat">
          <div className="k">Sessions</div>
          <div className="v">{issue.affected_sessions_estimate.toLocaleString()}</div>
        </div>
      </div>

      <div className="card">
        <div className="muted" style={{ marginBottom: 10 }}>Events over time</div>
        <ResponsiveContainer width="100%" height={220}>
          <LineChart data={data}>
            <CartesianGrid stroke="#272c3a" vertical={false} />
            <XAxis dataKey="t" stroke="#8b93a7" fontSize={11} />
            <YAxis stroke="#8b93a7" fontSize={11} />
            <Tooltip contentStyle={chartTip} />
            <Line type="monotone" dataKey="events" stroke="#f0506e" strokeWidth={2} dot={false} />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="card">
        <div className="muted" style={{ marginBottom: 12 }}>Latest events</div>
        <div className="stack" style={{ gap: 10 }}>
          {events.length === 0 && <div className="muted">No sampled events.</div>}
          {events.map((e) => (
            <EventRow key={e.event_id} e={e} serviceId={issue.service_id} issueId={issue.id} />
          ))}
        </div>
      </div>
    </div>
  );
}

// Split OTLP attributes into "known" buckets (http/browser/os) and the rest, so the
// context reads like Sentry's tags/request/user cards rather than one flat blob.
function partition(attrs: Record<string, string> | null) {
  const http: [string, string][] = [];
  const client: [string, string][] = [];
  const rest: [string, string][] = [];
  for (const [k, v] of Object.entries(attrs || {})) {
    if (k.startsWith("http.") || k.startsWith("url.")) http.push([k, v]);
    else if (k.startsWith("browser.") || k.startsWith("os.") || k.startsWith("device.") || k.startsWith("client.")) client.push([k, v]);
    else rest.push([k, v]);
  }
  return { http, client, rest };
}

function KV({ title, rows }: { title: string; rows: [string, string][] }) {
  if (rows.length === 0) return null;
  return (
    <div style={{ marginTop: 10 }}>
      <div className="muted" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: 0.5, marginBottom: 4 }}>{title}</div>
      <div className="code" style={{ fontSize: 12 }}>
        {rows.map(([k, v]) => (
          <div key={k}>
            <span style={{ color: "#8b93a7" }}>{k}</span>  {v}
          </div>
        ))}
      </div>
    </div>
  );
}

function EventRow({ e, serviceId, issueId }: { e: ErrorEvent; serviceId: number; issueId: number }) {
  const [open, setOpen] = useState(false);
  const [crumbs, setCrumbs] = useState<Breadcrumb[] | null>(null);
  const [sessionContext, setSessionContext] = useState<SessionContext | null>(null);
  const [contextLoading, setContextLoading] = useState(false);
  const { http, client, rest } = partition(e.attributes);

  const toggle = () => {
    const next = !open;
    setOpen(next);
    if (next && crumbs === null && e.session_id) {
      api
        .breadcrumbs(serviceId, e.session_id, e.timestamp)
        .then((r) => setCrumbs(r.breadcrumbs || []))
        .catch(() => setCrumbs([]));
    }
    if (next && sessionContext === null && !contextLoading) {
      setContextLoading(true);
      api
        .sessionContext(issueId, e.event_id)
        .then(setSessionContext)
        .catch(() => setSessionContext({
          status: "temporarily_unavailable",
          focused_at: e.timestamp,
          journey: [],
          network_failures: [],
          console_errors: [],
          exceptions: [],
          counts: {},
          truncated: false,
        }))
        .finally(() => setContextLoading(false));
    }
  };

  return (
    <div style={{ borderBottom: "1px solid var(--border)", paddingBottom: 10 }}>
      <div className="row spread" style={{ cursor: "pointer" }} onClick={toggle}>
        <div style={{ fontWeight: 600 }}>
          <span style={{ color: "#8b93a7", marginRight: 6 }}>{open ? "▾" : "▸"}</span>
          {e.exception_type}: {e.exception_message}
        </div>
        <div className="muted" style={{ fontSize: 12 }}>{new Date(e.timestamp).toLocaleString()}</div>
      </div>
      <div className="muted" style={{ fontSize: 12, marginTop: 4 }}>
        {e.environment && `env: ${e.environment}  `}
        {e.release && `release: ${e.release}  `}
        {e.user_id && `user: ${e.user_id}  `}
        {e.session_id && `session: ${e.session_id}`}
      </div>

      {e.stack_frames?.length > 0 && (
        <div className="code" style={{ marginTop: 8 }}>
          {e.stack_frames.map((f, i) => (
            <div key={i}>
              {f.function && <span style={{ color: "#7dd3fc" }}>{f.function}</span>}
              {f.function && " "}
              <span style={{ color: "#8b93a7" }}>{f.file}:{f.line}</span>
            </div>
          ))}
        </div>
      )}

      {open && (
        <div style={{ marginTop: 4 }}>
          <KV title="Request" rows={http} />
          <KV title="Client" rows={client} />
          <KV title="Tags & context" rows={rest} />
          {(e.trace_id || e.span_id) && (
            <KV title="Trace" rows={[["trace_id", e.trace_id], ["span_id", e.span_id]].filter(([, v]) => v) as [string, string][]} />
          )}

          <div style={{ marginTop: 10 }}>
            <div className="muted" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: 0.5, marginBottom: 4 }}>
              Breadcrumbs
            </div>
            {crumbs === null && <div className="muted" style={{ fontSize: 12 }}>Loading…</div>}
            {crumbs !== null && crumbs.length === 0 && (
              <div className="muted" style={{ fontSize: 12 }}>No prior events in this session.</div>
            )}
            {crumbs !== null && crumbs.length > 0 && (
              <div className="code" style={{ fontSize: 12 }}>
                {crumbs.map((b, i) => (
                  <div key={i} className="row" style={{ gap: 8 }}>
                    <span style={{ color: "#8b93a7", minWidth: 66 }}>
                      {new Date(b.timestamp).toLocaleTimeString()}
                    </span>
                    <span className={`badge lvl-${(b.severity_text || "info").toLowerCase()}`} style={{ fontSize: 10 }}>
                      {b.severity_text || "INFO"}
                    </span>
                    <span>{b.exception_type ? `${b.exception_type}: ${b.body}` : b.body}</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          <SessionContextView context={sessionContext} loading={contextLoading} />
        </div>
      )}
    </div>
  );
}

function SessionContextView({ context, loading }: { context: SessionContext | null; loading: boolean }) {
  if (loading) {
    return <div className="session-context muted" style={{ fontSize: 12 }}>Loading session context…</div>;
  }
  if (!context) return null;

  const messages: Record<SessionContext["status"], string> = {
    not_configured: "OpenReplay is not configured for this project.",
    missing_session: "This event does not contain a browser session ID.",
    recording_pending: "The recording is still being processed by OpenReplay.",
    temporarily_unavailable: "OpenReplay is temporarily unavailable. The rest of this event remains accessible.",
    ready: "",
  };
  if (context.status !== "ready") {
    return <div className="session-context muted" style={{ fontSize: 12 }}>{messages[context.status]}</div>;
  }

  return (
    <div className="session-context">
      <div className="row spread" style={{ marginBottom: 8 }}>
        <div>
          <div style={{ fontWeight: 700 }}>Session context</div>
          <div className="muted" style={{ fontSize: 12 }}>
            {context.session_id}
            {context.truncated && " · bounded view"}
          </div>
        </div>
        {context.replay_url && (
          <a className="replay-link" href={context.replay_url} target="_blank" rel="noreferrer">
            Watch replay ↗
          </a>
        )}
      </div>
      <ContextEvents title="User journey" events={context.journey} />
      <ContextEvents title="Failed requests" events={context.network_failures} />
      <ContextEvents title="Console errors" events={context.console_errors} />
      <ContextEvents title="Browser exceptions" events={context.exceptions} />
      {context.journey.length + context.network_failures.length + context.console_errors.length + context.exceptions.length === 0 && (
        <div className="muted" style={{ fontSize: 12 }}>No events were recorded in the error window.</div>
      )}
    </div>
  );
}

function ContextEvents({ title, events }: { title: string; events: SessionContextEvent[] }) {
  if (events.length === 0) return null;
  return (
    <div style={{ marginTop: 10 }}>
      <div className="muted" style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: 0.5 }}>
        {title}
      </div>
      {events.map((event) => (
        <div className="context-event" key={`${event.source_event_id}-${event.kind}`}>
          <span className="muted" style={{ fontSize: 11 }}>{new Date(event.timestamp).toLocaleTimeString()}</span>
          <span className="context-kind">{event.kind.replaceAll("_", " ")}</span>
          <span>
            {event.label}
            {event.count && event.count > 1 ? ` ×${event.count}` : ""}
          </span>
        </div>
      ))}
    </div>
  );
}
