import { useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../auth";
import { useOrg } from "../org";
import { api } from "../api";
import Modal from "../Modal";

export default function Shell() {
  const { user, logout } = useAuth();
  const { orgs, current, setCurrent, reload } = useOrg();
  const nav = useNavigate();
  const [creating, setCreating] = useState(false);
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");

  const closeModal = () => {
    setCreating(false);
    setName("");
    setErr("");
  };

  const createOrg = async () => {
    if (!name.trim() || busy) return;
    setBusy(true);
    setErr("");
    try {
      const r = await api.createOrg(name.trim());
      await reload();
      setCurrent(r.organization);
      closeModal();
      nav("/");
    } catch (e: any) {
      setErr(e.message || "Failed to create organization.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">
          errlog<span>.</span>
        </div>

        <div className="orgfield">
          <label className="field-label" htmlFor="orgswitch">
            Organization
          </label>
          <select
            id="orgswitch"
            className="orgswitch"
            value={current?.id ?? ""}
            onChange={(e) => {
              const o = orgs.find((x) => x.id === Number(e.target.value));
              if (o) {
                setCurrent(o);
                nav("/");
              }
            }}
          >
            {orgs.length === 0 && <option value="">No organization</option>}
            {orgs.map((o) => (
              <option key={o.id} value={o.id}>
                {o.name}
              </option>
            ))}
          </select>
        </div>

        <nav className="nav">
          <div className="nav-label">Menu</div>
          <NavLink to="/" end>
            <svg className="nav-ico" viewBox="0 0 24 24" aria-hidden="true">
              <rect x="3" y="3" width="7" height="7" rx="1.5" />
              <rect x="14" y="3" width="7" height="7" rx="1.5" />
              <rect x="3" y="14" width="7" height="7" rx="1.5" />
              <rect x="14" y="14" width="7" height="7" rx="1.5" />
            </svg>
            Projects
          </NavLink>
          <NavLink to="/members">
            <svg className="nav-ico" viewBox="0 0 24 24" aria-hidden="true">
              <circle cx="9" cy="8" r="3.2" />
              <path d="M3.5 20a5.5 5.5 0 0 1 11 0" />
              <path d="M16 5.2a3.2 3.2 0 0 1 0 6" />
              <path d="M17.5 14.4A5.5 5.5 0 0 1 20.5 19.4" />
            </svg>
            Members
          </NavLink>
        </nav>

        <div className="grow" />
        <button className="sidebar-cta" onClick={() => setCreating(true)}>
          <svg className="nav-ico" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 5v14M5 12h14" />
          </svg>
          New organization
        </button>

        <div className="userbox">
          <div className="avatar">{(user?.name || user?.email || "?").charAt(0).toUpperCase()}</div>
          <div className="userinfo">
            <div className="uname">{user?.name}</div>
            <div className="uemail">{user?.email}</div>
          </div>
          <button className="icon-btn" onClick={() => logout()} title="Sign out" aria-label="Sign out">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M15 4h3a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-3" />
              <path d="M10 17l-5-5 5-5" />
              <path d="M5 12h12" />
            </svg>
          </button>
        </div>
      </aside>

      <main className="main">
        <Outlet />
      </main>

      <Modal open={creating} title="Create organization" onClose={closeModal}>
        <form
          className="stack"
          style={{ gap: 14 }}
          onSubmit={(e) => {
            e.preventDefault();
            createOrg();
          }}
        >
          <input
            placeholder="Organization name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
          />
          {err && <div className="err">{err}</div>}
          <div className="row" style={{ justifyContent: "flex-end" }}>
            <button type="button" className="ghost" onClick={closeModal}>
              Cancel
            </button>
            <button type="submit" disabled={busy || !name.trim()}>
              {busy ? "Creating…" : "Create"}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
