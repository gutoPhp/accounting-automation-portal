import { useEffect, useState } from "react";

import { useHashRoute } from "./hooks/useHashRoute";
import { apiRequest } from "./lib/api";
import { AppShell } from "./components/layout/AppShell";
import { Brand } from "./components/ui/Brand";
import { LoginPage } from "./pages/LoginPage";

export function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const route = useHashRoute();

  useEffect(() => {
    apiRequest("/api/me")
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  async function handleLogout() {
    try {
      await apiRequest("/api/logout", { method: "POST" });
    } finally {
      setUser(null);
    }
  }

  if (loading) {
    return (
      <div className="splash">
        <Brand />
        <span>Preparando o portal…</span>
      </div>
    );
  }

  if (!user) {
    return <LoginPage onLogin={setUser} />;
  }

  return <AppShell user={user} route={route} onLogout={handleLogout} />;
}
