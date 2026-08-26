import { useEffect, useState } from "react";

import { Toast } from "../components/ui/Feedback";
import { Metric } from "../components/ui/Metric";
import { PageHeader } from "../components/ui/PageHeader";
import { apiRequest } from "../lib/api";

function daysUntil(date) {
  return Math.ceil((new Date(date) - new Date()) / 86_400_000);
}

function deadlineStatus(days) {
  if (days <= 30) {
    return "urgent";
  }
  if (days <= 60) {
    return "warning";
  }
  return "ok";
}

function CertificateRow({ certificate, onNotify }) {
  const remainingDays = daysUntil(certificate.expires_at);

  return (
    <tr>
      <td>
        <strong>{certificate.client_name}</strong>
      </td>
      <td>
        <code>{certificate.kind}</code>
      </td>
      <td>{new Date(certificate.expires_at).toLocaleDateString("pt-BR")}</td>
      <td>
        <span className={`status ${deadlineStatus(remainingDays)}`}>
          {remainingDays} dias
        </span>
      </td>
      <td>
        {certificate.last_notified_at ? (
          <>
            <strong>
              {new Date(certificate.last_notified_at).toLocaleDateString(
                "pt-BR",
              )}
            </strong>
            <small>{certificate.last_notified_to}</small>
          </>
        ) : (
          <span className="muted">Ainda não comunicado</span>
        )}
      </td>
      <td>
        <button
          className="ghost"
          disabled={remainingDays > 60}
          onClick={() => onNotify(certificate)}
        >
          {certificate.last_notified_at ? "Reavisar" : "Registrar aviso"}
        </button>
      </td>
    </tr>
  );
}

export function SC20Page() {
  const [certificates, setCertificates] = useState([]);
  const [schedule, setSchedule] = useState(null);
  const [toast, setToast] = useState(null);

  function loadCertificates() {
    return apiRequest("/api/sc-20/certificates").then(setCertificates);
  }

  useEffect(() => {
    loadCertificates();
    apiRequest("/api/sc-20/schedule").then(setSchedule);
  }, []);

  function showToast(type, text) {
    setToast({ type, text });
    window.setTimeout(() => setToast(null), 4000);
  }

  async function runScan() {
    try {
      const execution = await apiRequest("/api/sc-20/run", { method: "POST" });
      showToast("success", execution.summary);
    } catch (error) {
      showToast("error", error.message);
    }
  }

  async function notify(certificate) {
    try {
      await apiRequest("/api/sc-20/notify", {
        method: "POST",
        body: JSON.stringify({
          id: certificate.id,
          recipient: "responsavel@cliente.demo",
        }),
      });
      showToast("success", `Aviso de ${certificate.client_name} registrado.`);
      await loadCertificates();
    } catch (error) {
      showToast("error", error.message);
    }
  }

  const upTo30Days = certificates.filter(
    (certificate) => daysUntil(certificate.expires_at) <= 30,
  ).length;
  const from31To60Days = certificates.filter((certificate) => {
    const days = daysUntil(certificate.expires_at);
    return days > 30 && days <= 60;
  }).length;
  const outsideWindow = certificates.filter(
    (certificate) => daysUntil(certificate.expires_at) > 60,
  ).length;

  return (
    <>
      <Toast value={toast} />
      <PageHeader
        code="SC-20"
        title="Certificados digitais"
        subtitle="Antecedência para agir, sem repetir alertas que já foram tratados."
      >
        <button className="primary" onClick={runScan}>
          Executar varredura
        </button>
      </PageHeader>

      <section className="metrics compact-metrics">
        <Metric value={upTo30Days} label="Até 30 dias" tone="danger" />
        <Metric value={from31To60Days} label="De 31 a 60 dias" />
        <Metric value={outsideWindow} label="Fora da janela" />
      </section>

      {schedule && (
        <section className="panel schedule-panel">
          <div>
            <span className="status ok">Automático ativo</span>
            <strong>Job mensal de certificados</strong>
          </div>
          <p>
            Última execução automática:{" "}
            <strong>
              {schedule.last_run_at
                ? new Date(schedule.last_run_at).toLocaleString("pt-BR")
                : "Ainda não executado"}
            </strong>
          </p>
          <p>
            Próxima competência:{" "}
            <strong>
              {new Date(schedule.next_run_at).toLocaleDateString("pt-BR")}
            </strong>
          </p>
        </section>
      )}

      <section className="panel table-panel">
        <div className="panel-head">
          <div>
            <span className="kicker">MONITORAMENTO</span>
            <h2>Próximos vencimentos</h2>
          </div>
          <span className="tag">Janela de 60 dias</span>
        </div>

        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Cliente</th>
                <th>Tipo</th>
                <th>Validade</th>
                <th>Prazo</th>
                <th>Última comunicação</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {certificates.map((certificate) => (
                <CertificateRow
                  key={certificate.id}
                  certificate={certificate}
                  onNotify={notify}
                />
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </>
  );
}
