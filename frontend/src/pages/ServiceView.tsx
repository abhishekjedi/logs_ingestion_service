import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, type Issue, type OverviewPoint, type ReleaseHealth } from "../api";
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";

const hour = (t: string) => new Date(t).toLocaleString([], { month: "short", day: "numeric", hour: "2-digit" });
const chartTip = { background: "#1e222d", border: "1px solid #272c3a", borderRadius: 8, color: "#e6e9ef" };

export default function ServiceView() {
  const { serviceId } = useParams();
  const sid = Number(serviceId);
  const [tab, setTab] = useState<"issues" | "overview" | "health">("issues");

  return (
    <div>
      <Link to="/" className="muted" style={{ fontSize: 13 }}>
        ← Projects
      </Link>
      <h1 className="page-title" style={{ marginTop: 6 }}>
        Service #{sid}
      </h1>
      <div className="tabs">
        <a className={tab === "issues" ? "active" : ""} onClick={() => setTab("issues")}>
          Issues
        </a>
        <a className={tab === "overview" ? "active" : ""} onClick={() => setTab("overview")}>
          Overview
        </a>
        <a className={tab === "health" ? "active" : ""} onClick={() => setTab("health")}>
          Release health
        </a>
      </div>

      {tab === "issues" && <IssuesTab sid={sid} />}
      {tab === "overview" && <OverviewTab sid={sid} />}
      {tab === "health" && <HealthTab sid={sid} />}
    </div>
  );
}

function IssuesTab({ sid }: { sid: number }) {
  const [issues, setIssues] = useState<Issue[]>([]);
  const [sort, setSort] = useState("event_count");
  const nav = useNavigate();

  useEffect(() => {
    api.listIssues(sid, `?sort=${sort}&order=desc&limit=100`).then((r) => setIssues(r.issues || []));
  }, [sid, sort]);

  if (issues.length === 0) return <div className="empty">No issues yet — send some errors to this service.</div>;

  return (
    <table>
      <thead>
        <tr>
          <th>Issue</th>
          <th>Level</th>
          <th>Status</th>
          <th className="clickable" onClick={() => setSort("event_count")}>
            Events
          </th>
          <th>Users</th>
          <th className="clickable" onClick={() => setSort("last_seen")}>
            Last seen
          </th>
        </tr>
      </thead>
      <tbody>
        {issues.map((i) => (
          <tr key={i.id} className="clickable" onClick={() => nav(`/issues/${i.id}`)}>
            <td>
              <div style={{ fontWeight: 600 }}>{i.title || "(untitled)"}</div>
              {i.culprit && <div className="muted" style={{ fontSize: 12 }}>{i.culprit}</div>}
            </td>
            <td>
              <span className={`badge lvl-${i.level}`}>{i.level}</span>
            </td>
            <td>
              <span className={`badge st-${i.status}`}>{i.status}</span>
            </td>
            <td style={{ fontWeight: 700 }}>{i.event_count.toLocaleString()}</td>
            <td>{i.affected_users_estimate.toLocaleString()}</td>
            <td className="muted">{new Date(i.last_seen).toLocaleString()}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function OverviewTab({ sid }: { sid: number }) {
  const [points, setPoints] = useState<OverviewPoint[]>([]);
  useEffect(() => {
    api.serviceOverview(sid).then((r) => setPoints(r.points || []));
  }, [sid]);

  const totals = points.reduce(
    (a, p) => ({ events: a.events + p.events, users: Math.max(a.users, p.users), issues: Math.max(a.issues, p.issues) }),
    { events: 0, users: 0, issues: 0 }
  );
  const data = points.map((p) => ({ t: hour(p.timestamp), events: p.events, users: p.users }));

  return (
    <div className="stack">
      <div className="grid cols3">
        <div className="stat">
          <div className="k">Events (24h)</div>
          <div className="v">{totals.events.toLocaleString()}</div>
        </div>
        <div className="stat">
          <div className="k">Distinct issues</div>
          <div className="v">{totals.issues.toLocaleString()}</div>
        </div>
        <div className="stat">
          <div className="k">Affected users</div>
          <div className="v">{totals.users.toLocaleString()}</div>
        </div>
      </div>
      <div className="card">
        <div className="muted" style={{ marginBottom: 10 }}>Events over time</div>
        <ResponsiveContainer width="100%" height={260}>
          <LineChart data={data}>
            <CartesianGrid stroke="#272c3a" vertical={false} />
            <XAxis dataKey="t" stroke="#8b93a7" fontSize={11} />
            <YAxis stroke="#8b93a7" fontSize={11} />
            <Tooltip contentStyle={chartTip} />
            <Line type="monotone" dataKey="events" stroke="#6d5efc" strokeWidth={2} dot={false} />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

function HealthTab({ sid }: { sid: number }) {
  const [rows, setRows] = useState<ReleaseHealth[]>([]);
  useEffect(() => {
    api.releaseHealth(sid).then((r) => setRows(r.releases || []));
  }, [sid]);

  if (rows.length === 0) return <div className="empty">No session data — crash-free rate needs client-side session ids.</div>;

  return (
    <table>
      <thead>
        <tr>
          <th>Release</th>
          <th>Sessions</th>
          <th>Errored</th>
          <th>Crash-free</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r) => (
          <tr key={r.release}>
            <td style={{ fontWeight: 600 }}>{r.release || "(none)"}</td>
            <td>{r.sessions_total.toLocaleString()}</td>
            <td>{r.sessions_errored.toLocaleString()}</td>
            <td style={{ fontWeight: 700 }}>{(r.crash_free_rate * 100).toFixed(2)}%</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
