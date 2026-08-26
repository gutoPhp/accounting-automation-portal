import { useEffect, useState } from "react";

import { Toast } from "../components/ui/Feedback";
import { Metric } from "../components/ui/Metric";
import { PageHeader } from "../components/ui/PageHeader";
import { apiRequest } from "../lib/api";

function statusLabel(status) {
  const labels = {
    regular: "Regular",
    irregular: "Irregular",
    failed: "Consulta falhou",
  };
  return labels[status] || status;
}

function statusTone(status) {
  if (status === "failed") return "urgent";
  if (status === "irregular") return "warning";
  return "ok";
}

export function SC02Page() {
  const [checks, setChecks] = useState([]);
  const [schedule, setSchedule] = useState(null);
  const [simulateFailure, setSimulateFailure] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [toast, setToast] = useState(null);

  function loadData() {
    return Promise.all([
      apiRequest("/api/sc-02/checks"),
      apiRequest("/api/sc-02/schedule"),
    ]).then(([loadedChecks, loadedSchedule]) => {
      setChecks(Array.isArray(loadedChecks) ? loadedChecks : []);
      setSchedule(loadedSchedule);
    });
  }

  useEffect(() => {
    loadData();
  }, []);

  function showToast(type, text) {
    setToast({ type, text });
    window.setTimeout(() => setToast(null), 4000);
  }

  async function runScan() {
    setSubmitting(true);
    try {
      const execution = await apiRequest("/api/sc-02/run", {
        method: "POST",
        body: JSON.stringify({ simulate_failure: simulateFailure }),
      });
      showToast(
        execution.status === "success" ? "success" : "error",
        execution.summary,
      );
      await loadData();
    } catch (error) {
      showToast("error", error.message);
    } finally {
      setSubmitting(false);
    }
  }

  const regular = checks.filter((check) => check.status === "regular").length;
  const irregular = checks.filter(
    (check) => check.status === "irregular",
  ).length;
  const failed = checks.filter((check) => check.status === "failed").length;

  return (
    <>
      <Toast value={toast} />
      <PageHeader
        code="SC-02"
        title="Situação fiscal dos clientes"
        subtitle="Consultas multi-órgão com rastreabilidade de resultado, tentativa e falha."
      >
        <button className="primary" onClick={runScan} disabled={submitting}>
          {submitting ? "Consultando…" : "Executar consultas"}
        </button>
      </PageHeader>

      <section className="metrics compact-metrics">
        <Metric value={regular} label="Consultas regulares" />
        <Metric value={irregular} label="Irregularidades" tone="danger" />
        <Metric
          value={failed}
          label="Fontes com falha"
          tone={failed ? "danger" : ""}
        />
      </section>

      <section className="panel schedule-panel">
        <div>
          <span className="status ok">Automático ativo</span>
          <strong>Consulta fiscal mensal</strong>
        </div>
        {schedule && (
          <p>
            Última execução automática
            <strong>
              {schedule.last_run_at
                ? new Date(schedule.last_run_at).toLocaleString("pt-BR")
                : "Ainda não executada"}
            </strong>
          </p>
        )}
        <label className="check">
          <input
            type="checkbox"
            checked={simulateFailure}
            onChange={(event) => setSimulateFailure(event.target.checked)}
          />
          Simular falha no FGTS
        </label>
      </section>

      <section className="panel table-panel">
        <div className="panel-head">
          <div>
            <span className="kicker">ÚLTIMA RODADA</span>
            <h2>Resultado por cliente e órgão</h2>
          </div>
        </div>

        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Cliente</th>
                <th>Órgão</th>
                <th>Resultado</th>
                <th>Tentativas</th>
                <th>Consultado em</th>
                <th>Detalhe</th>
              </tr>
            </thead>
            <tbody>
              {checks.map((check) => (
                <tr key={check.id}>
                  <td>
                    <strong>{check.client_name}</strong>
                  </td>
                  <td>{check.agency}</td>
                  <td>
                    <span className={`status ${statusTone(check.status)}`}>
                      {statusLabel(check.status)}
                    </span>
                  </td>
                  <td>
                    <code>{check.attempts}</code>
                  </td>
                  <td>{new Date(check.checked_at).toLocaleString("pt-BR")}</td>
                  <td>{check.error || "Consulta concluída"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </>
  );
}
