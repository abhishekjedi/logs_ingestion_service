import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Project } from "../api";
import { useOrg } from "../org";

export default function Projects() {
  const { current, loading: orgLoading } = useOrg();
  const [projects, setProjects] = useState<Project[]>([]);
  const [name, setName] = useState("");
  const [adding, setAdding] = useState(false);

  const load = () => {
    if (current) api.listProjects(current.id).then((r) => setProjects(r.projects || []));
  };
  useEffect(load, [current]);

  const create = async () => {
    if (!name.trim() || !current) return;
    await api.createProject(current.id, name.trim());
    setName("");
    setAdding(false);
    load();
  };

  if (orgLoading) return <div className="muted">Loading…</div>;
  if (!current)
    return (
      <div className="empty">
        <h2>Welcome 👋</h2>
        <p>Create an organization from the sidebar to get started.</p>
      </div>
    );

  return (
    <div>
      <div className="row spread">
        <h1 className="page-title">Projects</h1>
        {adding ? (
          <div className="row">
            <input placeholder="Project name" value={name} onChange={(e) => setName(e.target.value)} autoFocus />
            <button onClick={create}>Create</button>
            <button className="ghost" onClick={() => setAdding(false)}>
              Cancel
            </button>
          </div>
        ) : (
          <button onClick={() => setAdding(true)}>+ New project</button>
        )}
      </div>

      {projects.length === 0 ? (
        <div className="empty">No projects yet. Create one to register a service.</div>
      ) : (
        <div className="grid cols3">
          {projects.map((p) => (
            <Link key={p.id} to={`/projects/${p.id}`} className="card">
              <div style={{ fontWeight: 700, fontSize: 16 }}>{p.name}</div>
              <div className="muted" style={{ fontSize: 13, marginTop: 6 }}>
                Project #{p.id}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
