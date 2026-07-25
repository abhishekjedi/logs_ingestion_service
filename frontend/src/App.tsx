import { Routes, Route, Navigate } from "react-router-dom";
import { useAuth } from "./auth";
import { OrgProvider } from "./org";
import Login from "./pages/Login";
import Shell from "./pages/Shell";
import Projects from "./pages/Projects";
import Services from "./pages/Services";
import ServiceView from "./pages/ServiceView";
import IssueDetail from "./pages/IssueDetail";
import Members from "./pages/Members";

export default function App() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="login">
        <div className="muted">Loading…</div>
      </div>
    );
  }

  if (!user) {
    return (
      <Routes>
        <Route path="*" element={<Login />} />
      </Routes>
    );
  }

  return (
    <OrgProvider>
      <Routes>
        <Route element={<Shell />}>
          <Route index element={<Projects />} />
          <Route path="projects/:projectId" element={<Services />} />
          <Route path="services/:serviceId" element={<ServiceView />} />
          <Route path="issues/:issueId" element={<IssueDetail />} />
          <Route path="members" element={<Members />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </OrgProvider>
  );
}
