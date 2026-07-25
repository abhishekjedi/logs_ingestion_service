import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { api, type Org } from "./api";

type OrgCtx = {
  orgs: Org[];
  current: Org | null;
  loading: boolean;
  setCurrent: (o: Org) => void;
  reload: () => Promise<void>;
};

const Ctx = createContext<OrgCtx>(null!);

export function OrgProvider({ children }: { children: ReactNode }) {
  const [orgs, setOrgs] = useState<Org[]>([]);
  const [current, setCurrentState] = useState<Org | null>(null);
  const [loading, setLoading] = useState(true);

  const reload = async () => {
    const r = await api.listOrgs();
    setOrgs(r.organizations);
    const saved = localStorage.getItem("orgId");
    const found = r.organizations.find((o) => String(o.id) === saved) || r.organizations[0] || null;
    setCurrentState(found);
    setLoading(false);
  };

  useEffect(() => {
    reload();
  }, []);

  const setCurrent = (o: Org) => {
    localStorage.setItem("orgId", String(o.id));
    setCurrentState(o);
  };

  return <Ctx.Provider value={{ orgs, current, loading, setCurrent, reload }}>{children}</Ctx.Provider>;
}

export const useOrg = () => useContext(Ctx);
