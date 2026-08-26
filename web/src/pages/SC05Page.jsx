import { useEffect, useState } from "react";

import { EmptyState, Toast } from "../components/ui/Feedback";
import { PageHeader } from "../components/ui/PageHeader";
import { apiRequest } from "../lib/api";

function OperationChoice({ value, selected, title, description, onSelect }) {
  return (
    <button
      type="button"
      className={selected ? "selected" : ""}
      onClick={() => onSelect(value)}
    >
      <strong>{title}</strong>
      <span>{description}</span>
    </button>
  );
}

function ExecutionTimeline({ result }) {
  if (!result) {
    return (
      <EmptyState>
        Execute uma operação para acompanhar cada sistema.
      </EmptyState>
    );
  }

  return (
    <div className="timeline">
      {result.actions.map((action) => (
        <div className={`timeline-item ${action.status}`} key={action.system}>
          <span>{action.status === "success" ? "✓" : "!"}</span>
          <div>
            <strong>{action.system}</strong>
            <p>{action.action}</p>
            <small>{action.detail}</small>
          </div>
        </div>
      ))}
      {result.error && <div className="alert error">{result.error}</div>}
    </div>
  );
}

function SystemsComparison({ before, after }) {
  if (!before.length) {
    return null;
  }

  return (
    <section className="panel systems-panel">
      <div className="panel-head">
        <div>
          <span className="kicker">FRONTEIRAS MOCKADAS</span>
          <h2>Estado nos sistemas simulados</h2>
        </div>
      </div>

      <div className="systems-grid">
        {before.map((system, index) => {
          const current = after[index] || system;
          const changed = system.state !== current.state;

          return (
            <article className="system-card" key={system.system}>
              <strong>{system.system}</strong>
              <small>{system.detail}</small>
              <div className="state-change">
                <span>
                  Antes
                  <strong>{system.state}</strong>
                </span>
                <b>{changed ? "→" : "="}</b>
                <span className={changed ? "changed" : ""}>
                  Depois
                  <strong>{current.state}</strong>
                </span>
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}

export function SC05Page() {
  const [clients, setClients] = useState([]);
  const [clientID, setClientID] = useState("");
  const [action, setAction] = useState("block");
  const [simulateFailure, setSimulateFailure] = useState(false);
  const [result, setResult] = useState(null);
  const [systemsBefore, setSystemsBefore] = useState([]);
  const [systemsAfter, setSystemsAfter] = useState([]);
  const [submitting, setSubmitting] = useState(false);
  const [toast, setToast] = useState(null);

  function loadClients() {
    return apiRequest("/api/clients").then((data) => {
      setClients(data);
      setClientID((current) => current || data[0]?.id || "");
    });
  }

  useEffect(() => {
    loadClients();
  }, []);

  useEffect(() => {
    if (!clientID) return;

    apiRequest(`/api/sc-05/systems?client_id=${clientID}`).then((systems) => {
      setSystemsBefore(systems);
      setSystemsAfter(systems);
    });
  }, [clientID]);

  function showToast(type, text) {
    setToast({ type, text });
    window.setTimeout(() => setToast(null), 4000);
  }

  async function runAutomation() {
    setSubmitting(true);
    setResult(null);

    try {
      const before = await apiRequest(
        `/api/sc-05/systems?client_id=${clientID}`,
      );
      const execution = await apiRequest("/api/sc-05/run", {
        method: "POST",
        body: JSON.stringify({
          client_id: clientID,
          action,
          simulate_failure: simulateFailure,
        }),
      });
      const after = await apiRequest(
        `/api/sc-05/systems?client_id=${clientID}`,
      );

      setSystemsBefore(before);
      setSystemsAfter(after);
      setResult(execution);
      showToast(
        execution.status === "success" ? "success" : "error",
        execution.summary,
      );
      await loadClients();
    } catch (error) {
      showToast("error", error.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <Toast value={toast} />
      <PageHeader
        code="SC-05"
        title="Bloqueio de inadimplentes"
        subtitle="Uma operação coordenada, reversível e rastreável em todos os sistemas."
      />

      <div className="two-columns">
        <section className="panel">
          <div className="panel-head">
            <div>
              <span className="kicker">NOVA EXECUÇÃO</span>
              <h2>Configurar operação</h2>
            </div>
            <span className="tag">Sob demanda</span>
          </div>

          <label>
            Cliente
            <select
              value={clientID}
              onChange={(event) => setClientID(event.target.value)}
            >
              {clients.map((client) => (
                <option key={client.id} value={client.id}>
                  {client.name} · {client.blocked ? "Bloqueado" : "Ativo"}
                </option>
              ))}
            </select>
          </label>

          <div className="choice">
            <OperationChoice
              value="block"
              selected={action === "block"}
              title="Bloquear"
              description="Suspender acessos e preservar tarefas"
              onSelect={setAction}
            />
            <OperationChoice
              value="unblock"
              selected={action === "unblock"}
              title="Desbloquear"
              description="Restaurar acessos e responsáveis"
              onSelect={setAction}
            />
          </div>

          <label className="check">
            <input
              type="checkbox"
              checked={simulateFailure}
              onChange={(event) => setSimulateFailure(event.target.checked)}
            />
            Simular indisponibilidade para demonstrar tratamento de falha
          </label>

          <button
            className="primary"
            onClick={runAutomation}
            disabled={submitting || !clientID}
          >
            {submitting ? "Executando…" : "Executar sequência"}
          </button>
        </section>

        <section className="panel">
          <div className="panel-head">
            <div>
              <span className="kicker">RESULTADO</span>
              <h2>Etapas da execução</h2>
            </div>
          </div>
          <ExecutionTimeline result={result} />
        </section>
      </div>

      <SystemsComparison before={systemsBefore} after={systemsAfter} />
    </>
  );
}
