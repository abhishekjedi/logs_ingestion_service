import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type Service, type CreatedService, type ReplayIntegration } from "../api";

export default function Services() {
  const { projectId } = useParams();
  const pid = Number(projectId);
  const [services, setServices] = useState<Service[]>([]);
  const [name, setName] = useState("");
  const [adding, setAdding] = useState(false);
  const [created, setCreated] = useState<CreatedService | null>(null);
  const [integrations, setIntegrations] = useState<ReplayIntegration[]>([]);
  const [canManageReplay, setCanManageReplay] = useState(false);
  const [replayProjectKey, setReplayProjectKey] = useState("");
  const [replayAPIKey, setReplayAPIKey] = useState("");
  const [replayAPIBase, setReplayAPIBase] = useState("https://api.openreplay.com/v2");
  const [replayIngest, setReplayIngest] = useState("https://api.openreplay.com/ingest");
  const [replayBusy, setReplayBusy] = useState(false);
  const [replayError, setReplayError] = useState("");

  const load = useCallback(() => {
    api.listServices(pid).then((r) => setServices(r.services || []));
    api.listReplayIntegrations(pid).then((r) => {
      setIntegrations(r.integrations || []);
      setCanManageReplay(r.can_manage);
    });
  }, [pid]);
  useEffect(() => {
    load();
  }, [load]);

  const create = async () => {
    if (!name.trim()) return;
    const r = await api.createService(pid, name.trim());
    setCreated(r);
    setName("");
    setAdding(false);
    load();
  };

  const saveReplay = async () => {
    if (!replayProjectKey.trim() || replayBusy) return;
    setReplayBusy(true);
    setReplayError("");
    try {
      await api.saveReplayIntegration(pid, {
        external_project_key: replayProjectKey.trim(),
        api_base_url: replayAPIBase.trim(),
        ingest_point: replayIngest.trim(),
        organization_api_key: replayAPIKey.trim() || undefined,
        enabled: true,
      });
      setReplayAPIKey("");
      load();
    } catch (error) {
      setReplayError(error instanceof Error ? error.message : "Could not connect OpenReplay.");
      load();
    } finally {
      setReplayBusy(false);
    }
  };

  const editReplay = (integration: ReplayIntegration) => {
    setReplayProjectKey(integration.external_project_key);
    setReplayAPIBase(integration.api_base_url);
    setReplayIngest(integration.ingest_point);
    setReplayAPIKey("");
    setReplayError("");
  };

  const removeReplay = async (integration: ReplayIntegration) => {
    await api.deleteReplayIntegration(pid, integration.external_project_key);
    if (replayProjectKey === integration.external_project_key) setReplayProjectKey("");
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

      <div className="card stack" style={{ marginBottom: 18 }}>
        <div>
          <div style={{ fontWeight: 700, fontSize: 16 }}>OpenReplay</div>
          <div className="muted" style={{ fontSize: 13, marginTop: 4 }}>
            Connect browser sessions to backend errors and compile them into a bounded debugging timeline.
          </div>
        </div>

        {integrations.length > 0 && (
          <div className="stack" style={{ gap: 8 }}>
            {integrations.map((integration) => (
              <div className="row spread replay-integration" key={integration.id}>
                <div>
                  <div style={{ fontWeight: 600 }}>{integration.external_project_key}</div>
                  <div className="muted" style={{ fontSize: 12 }}>
                    {integration.enabled ? "Connected" : "Disabled"}
                    {integration.last_validated_at && ` · validated ${new Date(integration.last_validated_at).toLocaleString()}`}
                  </div>
                </div>
                {canManageReplay && (
                  <div className="row" style={{ gap: 8 }}>
                    <button className="ghost" onClick={() => editReplay(integration)}>Edit</button>
                    <button className="ghost danger-ghost" onClick={() => removeReplay(integration)}>Disconnect</button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {canManageReplay ? (
          <>
            <div className="integration-form">
              <label>
                <span>OpenReplay project key</span>
                <input
                  value={replayProjectKey}
                  onChange={(event) => setReplayProjectKey(event.target.value)}
                  placeholder="Project key"
                />
              </label>
              <label>
                <span>Organization API key</span>
                <input
                  type="password"
                  value={replayAPIKey}
                  onChange={(event) => setReplayAPIKey(event.target.value)}
                  placeholder="Leave empty to keep the saved key"
                />
              </label>
              <label>
                <span>API base URL</span>
                <input value={replayAPIBase} onChange={(event) => setReplayAPIBase(event.target.value)} />
              </label>
              <label>
                <span>Tracker ingest point</span>
                <input value={replayIngest} onChange={(event) => setReplayIngest(event.target.value)} />
              </label>
            </div>
            {replayError && <div className="err">{replayError}</div>}
            <div>
              <button disabled={replayBusy || !replayProjectKey.trim()} onClick={saveReplay}>
                {replayBusy ? "Validating…" : "Connect OpenReplay"}
              </button>
            </div>
          </>
        ) : (
          <div className="muted" style={{ fontSize: 12 }}>Only project owners and admins can change this integration.</div>
        )}
      </div>

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
