import { Brand } from "../ui/Brand";
import { DashboardPage } from "../../pages/DashboardPage";
import { HistoryPage } from "../../pages/HistoryPage";
import { SC02Page } from "../../pages/SC02Page";
import { SC05Page } from "../../pages/SC05Page";
import { SC06Page } from "../../pages/SC06Page";
import { SC20Page } from "../../pages/SC20Page";

function buildNavigation(user) {
  const modules = user.modules.map((code) => ({
    id: code.toLowerCase(),
    label: code,
    icon:
      code === "SC-02"
        ? "◎"
        : code === "SC-05"
          ? "↯"
          : code === "SC-06"
            ? "✓"
            : "◷",
  }));

  return [
    { id: "home", label: "Visão geral", icon: "⌂" },
    ...modules,
    { id: "history", label: "Execuções", icon: "≡" },
  ];
}

function ActivePage({ route, user }) {
  switch (route) {
    case "sc-02":
      return <SC02Page />;
    case "sc-05":
      return <SC05Page />;
    case "sc-06":
      return <SC06Page />;
    case "sc-20":
      return <SC20Page />;
    case "history":
      return <HistoryPage />;
    default:
      return <DashboardPage user={user} />;
  }
}

export function AppShell({ user, route, onLogout }) {
  const navigation = buildNavigation(user);

  return (
    <div className="app-shell">
      <aside>
        <Brand compact />

        <nav>
          {navigation.map((item) => (
            <a
              key={item.id}
              className={route === item.id ? "active" : ""}
              href={`#${item.id}`}
            >
              <span>{item.icon}</span>
              {item.label}
            </a>
          ))}
        </nav>

        <div className="profile">
          <div className="avatar">{user.name[0]}</div>
          <div>
            <strong>{user.name}</strong>
            <span>{user.role === "admin" ? "Administrador" : "Operador"}</span>
          </div>
          <button onClick={onLogout} title="Sair">
            ↪
          </button>
        </div>
      </aside>

      <main className="content">
        <ActivePage route={route} user={user} />
      </main>
    </div>
  );
}
