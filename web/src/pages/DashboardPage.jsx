import { useEffect, useState } from "react";

import { Metric } from "../components/ui/Metric";
import { PageHeader } from "../components/ui/PageHeader";
import { apiRequest } from "../lib/api";

function ModuleCard({ module }) {
  return (
    <a className="module-card" href={`#${module.code.toLowerCase()}`}>
      <div className="card-top">
        <span className="module-code">{module.code}</span>
        <span className="nature">{module.nature}</span>
      </div>
      <h3>{module.name}</h3>
      <p>{module.description}</p>
      <span className="open">Abrir módulo →</span>
    </a>
  );
}

export function DashboardPage({ user }) {
  const [dashboard, setDashboard] = useState(null);

  useEffect(() => {
    apiRequest("/api/dashboard").then(setDashboard);
  }, []);

  if (!dashboard) {
    return <div className="loading">Carregando visão geral…</div>;
  }

  return (
    <>
      <PageHeader
        title={`Olá, ${user.name.split(" ")[0]}.`}
        subtitle="Acompanhe os processos e acesse as automações disponíveis para o seu perfil."
      />

      <section className="metrics">
        <Metric
          value={dashboard.metrics.clients}
          label="Clientes monitorados"
        />
        <Metric
          value={dashboard.metrics.executions}
          label="Execuções registradas"
        />
        <Metric
          value={dashboard.metrics.failures}
          label="Falhas para analisar"
          tone={dashboard.metrics.failures ? "danger" : ""}
        />
        <Metric
          value={dashboard.metrics.certificates}
          label="Certificados na base"
        />
      </section>

      <div className="section-title">
        <div>
          <span className="kicker">AUTOMAÇÕES</span>
          <h2>Módulos operacionais</h2>
        </div>
        <span>{dashboard.modules.length} disponíveis</span>
      </div>

      <section className="module-grid">
        {dashboard.modules.map((module) => (
          <ModuleCard key={module.code} module={module} />
        ))}
      </section>
    </>
  );
}
