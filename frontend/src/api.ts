// Typed-ish client for the error-logging API. Cookies (session) ride along via
// credentials:include; the Vite proxy keeps it same-origin in dev.

export type User = { id: number; email: string; name: string; avatar_url: string };
export type Org = { id: number; name: string; slug: string };
export type Member = { id: number; email: string; role: string; status: string; user_id?: number };
export type Project = { id: number; name: string; org_id?: number; created_at: string };
export type Service = { id: number; name: string; public_id: string; api_key_prefix: string; created_at: string };
export type CreatedService = { service: Service; api_key: string; ingest_url: string };
export type Issue = {
  id: number; service_id: number; title: string; culprit: string; level: string; status: string;
  event_count: number; affected_users_estimate: number; affected_sessions_estimate: number;
  first_seen: string; last_seen: string;
};
export type TimePoint = { timestamp: string; events: number; users: number };
export type OverviewPoint = { timestamp: string; events: number; issues: number; users: number };
export type ReleaseHealth = { release: string; sessions_total: number; sessions_errored: number; crash_free_rate: number };
export type ErrorEvent = {
  event_id: string; timestamp: string; severity_text: string; exception_type: string;
  exception_message: string; user_id: string; session_id: string; environment: string; release: string;
  trace_id: string; span_id: string;
  stack_frames: { file: string; function: string; line: number }[];
  attributes: Record<string, string> | null;
  resource_attributes: Record<string, string> | null;
};
export type Breadcrumb = { timestamp: string; severity_text: string; body: string; exception_type: string };

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch("/api" + path, {
    method,
    credentials: "include",
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new ApiError(res.status, (data as any).error || res.statusText);
  }
  if (res.status === 204) return null as T;
  return res.json();
}

export const api = {
  me: () => req<{ user: User }>("GET", "/auth/me"),
  logout: () => req<void>("POST", "/auth/logout"),

  listOrgs: () => req<{ organizations: Org[] }>("GET", "/orgs"),
  createOrg: (name: string) => req<{ organization: Org }>("POST", "/orgs", { name }),
  listMembers: (orgId: number) => req<{ members: Member[] }>("GET", `/orgs/${orgId}/members`),
  inviteMember: (orgId: number, email: string, role: string) =>
    req<{ member: Member }>("POST", `/orgs/${orgId}/members`, { email, role }),

  listProjects: (orgId: number) => req<{ projects: Project[] }>("GET", `/orgs/${orgId}/projects`),
  createProject: (orgId: number, name: string) => req<{ project: Project }>("POST", `/orgs/${orgId}/projects`, { name }),
  getProject: (id: number) => req<{ project: Project }>("GET", `/projects/${id}`),

  listServices: (projectId: number) => req<{ services: Service[] }>("GET", `/projects/${projectId}/services`),
  createService: (projectId: number, name: string) => req<CreatedService>("POST", `/projects/${projectId}/services`, { name }),

  listIssues: (serviceId: number, q = "") => req<{ issues: Issue[]; total: number }>("GET", `/services/${serviceId}/issues${q}`),
  getIssue: (id: number) => req<{ issue: Issue }>("GET", `/issues/${id}`),
  issueTimeseries: (id: number) => req<{ points: TimePoint[] }>("GET", `/issues/${id}/timeseries`),
  issueEvents: (id: number) => req<{ events: ErrorEvent[] }>("GET", `/issues/${id}/events`),

  serviceOverview: (id: number) => req<{ points: OverviewPoint[] }>("GET", `/services/${id}/overview`),
  releaseHealth: (id: number) => req<{ releases: ReleaseHealth[] }>("GET", `/services/${id}/release-health`),
  breadcrumbs: (serviceId: number, sessionId: string, before: string) =>
    req<{ breadcrumbs: Breadcrumb[] }>(
      "GET",
      `/services/${serviceId}/breadcrumbs?session_id=${encodeURIComponent(sessionId)}&before=${encodeURIComponent(before)}`,
    ),
};
