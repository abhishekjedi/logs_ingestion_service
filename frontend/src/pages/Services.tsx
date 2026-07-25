import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type Service, type CreatedService } from "../api";

export default function Services() {
  const { projectId } = useParams();
  const pid = Number(projectId);
  const [services, setServices] = useState<Service[]>([]);
  const [name, setName] = useState("");
  const [adding, setAdding] = useState(false);
  const [created, setCreated] = useState<CreatedService | null>(null);

  const load = () => api.listServices(pid).then((r) => setServices(r.services || []));
  useEffect(() => {
    load();
  }, [pid]);

  const create = async () => {
    if (!name.trim()) return;
    const r = await api.createService(pid, name.trim());
    setCreated(r);
    setName("");
    setAdding(false);
    load();
  };

  return (
    <div>
      <div className="row spread">
        <div>
          <Link to="/" className="muted" style={{ fontSize: 13 }}>
            ← Projects
          </Link>
          <h1 className="page-title" style={{ marginTop: 6 }}>
            Services
          </h1>
        </div>
        {adding ? (
          <div className="row">
            <input placeholder="Service name" value={name} onChange={(e) => setName(e.target.value)} autoFocus />
            <button onClick={create}>Create</button>
            <button className="ghost" onClick={() => setAdding(false)}>
              Cancel
            </button>
          </div>
        ) : (
          <button onClick={() => setAdding(true)}>+ New service</button>
        )}
      </div>

      {created && (
        <div className="card stack" style={{ borderColor: "var(--accent)", marginBottom: 18 }}>
          <div style={{ fontWeight: 700 }}>Service “{created.service.name}” created — copy your API key now (shown once):</div>
          <div className="code">{created.api_key}</div>
          <div className="muted" style={{ fontSize: 13 }}>Ingest URL</div>
          <div className="code">{created.ingest_url}</div>
          <div>
            <button className="ghost" onClick={() => setCreated(null)}>
              Done
            </button>
          </div>
        </div>
      )}

      {services.length === 0 ? (
        <div className="empty">No services yet. Create one to get an ingest URL + API key.</div>
      ) : (
        <div className="grid cols3">
          {services.map((s) => (
            <Link key={s.id} to={`/services/${s.id}`} className="card">
              <div style={{ fontWeight: 700, fontSize: 16 }}>{s.name}</div>
              <div className="muted" style={{ fontSize: 13, marginTop: 6 }}>
                key {s.api_key_prefix}…
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
