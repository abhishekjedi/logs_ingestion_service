import { useEffect, useState } from "react";
import { api, type Member } from "../api";
import { useOrg } from "../org";

export default function Members() {
  const { current } = useOrg();
  const [members, setMembers] = useState<Member[]>([]);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [err, setErr] = useState("");

  const load = () => {
    if (current) api.listMembers(current.id).then((r) => setMembers(r.members || [])).catch(() => {});
  };
  useEffect(load, [current]);

  const invite = async () => {
    if (!current || !email.trim()) return;
    setErr("");
    try {
      await api.inviteMember(current.id, email.trim(), role);
      setEmail("");
      load();
    } catch (e: any) {
      setErr(e.status === 403 ? "Only owners/admins can invite members." : e.message);
    }
  };

  if (!current) return <div className="empty">Select an organization.</div>;

  return (
    <div>
      <h1 className="page-title">Members · {current.name}</h1>

      <div className="card stack" style={{ marginBottom: 18 }}>
        <div style={{ fontWeight: 600 }}>Invite a teammate</div>
        <div className="row">
          <input className="grow" placeholder="teammate@company.com" value={email} onChange={(e) => setEmail(e.target.value)} type="email" />
          <select value={role} onChange={(e) => setRole(e.target.value)}>
            <option value="member">member</option>
            <option value="admin">admin</option>
          </select>
          <button onClick={invite}>Invite</button>
        </div>
        {err && <div className="err">{err}</div>}
        <div className="muted" style={{ fontSize: 12 }}>
          They join instantly if they already have an account, otherwise on first login.
        </div>
      </div>

      <table>
        <thead>
          <tr>
            <th>Email</th>
            <th>Role</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {members.map((m) => (
            <tr key={m.id}>
              <td style={{ fontWeight: 600 }}>{m.email}</td>
              <td>
                <span className="badge lvl-info">{m.role}</span>
              </td>
              <td>
                <span className={`badge ${m.status === "active" ? "st-resolved" : "st-regressed"}`}>{m.status}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
