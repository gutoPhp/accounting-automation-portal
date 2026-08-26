import { useEffect, useState } from "react";

import { EmptyState } from "../components/ui/Feedback";
import { PageHeader } from "../components/ui/PageHeader";
import { apiRequest } from "../lib/api";

function ExecutionRow({ execution }) {
  const succeeded = execution.status === "success";

  return (
    <tr>
      <td>
        <code>{execution.id}</code>
      </td>
      <td>
        <span className="module-code">{execution.module}</span>
      </td>
      <td>{new Date(execution.started_at).toLocaleString("pt-BR")}</td>
      <td>{execution.trigger === "manual" ? "Manual" : "Agendado"}</td>
      <td>{execution.triggered_by}</td>
      <td>
        <span className={`status ${succeeded ? "ok" : "urgent"}`}>
          {succeeded ? "Sucesso" : "Falhou"}
        </span>
        <small>{execution.summary}</small>
      </td>
    </tr>
  );
}

export function HistoryPage() {
  const [executions, setExecutions] = useState([]);

  useEffect(() => {
    apiRequest("/api/executions").then(setExecutions);
  }, []);

  return (
    <>
      <PageHeader
        title="Histórico de execuções"
        subtitle="Auditoria centralizada de cada disparo, resultado e falha operacional."
      />

      <section className="panel table-panel">
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Execução</th>
                <th>Módulo</th>
                <th>Início</th>
                <th>Disparo</th>
                <th>Responsável</th>
                <th>Resultado</th>
              </tr>
            </thead>
            <tbody>
              {executions.map((execution) => (
                <ExecutionRow key={execution.id} execution={execution} />
              ))}
              {!executions.length && (
                <tr>
                  <td colSpan="6">
                    <EmptyState>Nenhuma execução registrada.</EmptyState>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>
    </>
  );
}
