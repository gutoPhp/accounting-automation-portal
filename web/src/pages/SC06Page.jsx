import { useEffect, useMemo, useState } from "react";

import { Toast } from "../components/ui/Feedback";
import { PageHeader } from "../components/ui/PageHeader";
import { apiRequest } from "../lib/api";

const initialAnswers = {
  client_name: "",
  service_type: "abertura",
  state: "AL",
  partner_name: "",
  partner_marital_status: "solteiro",
};

const fieldLabels = {
  client_name: "Nome do cliente",
  service_type: "Tipo de serviço",
  state: "UF da sede",
  partner_name: "Nome do sócio",
  partner_marital_status: "Estado civil",
  foreign_state_registration: "Inscrição estadual na UF",
  foreign_state_city: "Município da sede",
  marriage_regime: "Regime de casamento",
  change_description: "Alteração solicitada",
};

function ruleMatches(rule, answers) {
  const actual = answers[rule.field]?.toLowerCase();
  const expected = rule.value.toLowerCase();
  return rule.operator === "eq"
    ? actual === expected
    : Boolean(actual) && actual !== expected;
}

function TextField({ name, value, missing, onChange }) {
  return (
    <label>
      {fieldLabels[name]}
      <input
        className={missing ? "invalid" : ""}
        value={value}
        onChange={(event) => onChange(name, event.target.value)}
      />
      {missing && (
        <small className="field-error">Campo obrigatório neste caso</small>
      )}
    </label>
  );
}

function RulesPanel({ rules, briefings }) {
  return (
    <section className="panel">
      <div className="panel-head">
        <div>
          <span className="kicker">REGRAS ATIVAS</span>
          <h2>Lógica aplicada</h2>
        </div>
      </div>

      <div className="rule-list">
        {rules.map((rule) => (
          <div key={rule.id}>
            <span>SE</span>
            <p>
              <strong>{fieldLabels[rule.field]}</strong>{" "}
              {rule.operator === "eq" ? "for" : "não for"} “{rule.value}”
            </p>
            <small>
              Exige:{" "}
              {rule.requires.map((field) => fieldLabels[field]).join(", ")}
            </small>
          </div>
        ))}
      </div>

      <div className="mini-history">
        <h3>Briefings concluídos</h3>
        {briefings.length ? (
          briefings.slice(0, 3).map((briefing) => (
            <p key={briefing.id}>
              <strong>{briefing.answers.client_name}</strong>
              <span>
                {new Date(briefing.created_at).toLocaleString("pt-BR")}
              </span>
            </p>
          ))
        ) : (
          <small>Nenhum briefing concluído.</small>
        )}
      </div>
    </section>
  );
}

export function SC06Page() {
  const [answers, setAnswers] = useState(initialAnswers);
  const [rules, setRules] = useState([]);
  const [briefings, setBriefings] = useState([]);
  const [missingFields, setMissingFields] = useState([]);
  const [toast, setToast] = useState(null);

  function loadData() {
    return Promise.all([
      apiRequest("/api/sc-06/rules"),
      apiRequest("/api/sc-06/briefings"),
    ]).then(([loadedRules, loadedBriefings]) => {
      setRules(Array.isArray(loadedRules) ? loadedRules : []);
      setBriefings(Array.isArray(loadedBriefings) ? loadedBriefings : []);
    });
  }

  useEffect(() => {
    loadData();
  }, []);

  const conditionalFields = useMemo(
    () =>
      rules
        .filter((rule) => ruleMatches(rule, answers))
        .flatMap((rule) => rule.requires),
    [rules, answers],
  );

  function updateAnswer(field, value) {
    setAnswers((current) => ({ ...current, [field]: value }));
  }

  function showToast(type, text) {
    setToast({ type, text });
    window.setTimeout(() => setToast(null), 4000);
  }

  async function handleSubmit(event) {
    event.preventDefault();
    setMissingFields([]);

    try {
      await apiRequest("/api/sc-06/briefings", {
        method: "POST",
        body: JSON.stringify({ answers }),
      });
      showToast("success", "Briefing completo e salvo.");
      setAnswers(initialAnswers);
      await loadData();
    } catch (error) {
      setMissingFields(error.data?.missing || []);
      showToast("error", error.message);
    }
  }

  return (
    <>
      <Toast value={toast} />
      <PageHeader
        code="SC-06"
        title="Briefing societário"
        subtitle="Perguntas que se adaptam ao caso e só concluem quando a informação está completa."
      />

      <div className="two-columns wide-left">
        <form className="panel form-grid" onSubmit={handleSubmit}>
          <div className="panel-head full">
            <div>
              <span className="kicker">NOVO BRIEFING</span>
              <h2>Dados do processo</h2>
            </div>
            <span className="tag">Regras configuráveis</span>
          </div>

          <TextField
            name="client_name"
            value={answers.client_name}
            missing={missingFields.includes("client_name")}
            onChange={updateAnswer}
          />

          <label>
            Tipo de serviço
            <select
              value={answers.service_type}
              onChange={(event) =>
                updateAnswer("service_type", event.target.value)
              }
            >
              <option value="abertura">Abertura de empresa</option>
              <option value="alteracao">Alteração contratual</option>
            </select>
          </label>

          <label>
            UF da sede
            <select
              value={answers.state}
              onChange={(event) => updateAnswer("state", event.target.value)}
            >
              {["AL", "PE", "SE", "BA", "SP"].map((state) => (
                <option key={state}>{state}</option>
              ))}
            </select>
          </label>

          <TextField
            name="partner_name"
            value={answers.partner_name}
            missing={missingFields.includes("partner_name")}
            onChange={updateAnswer}
          />

          <label>
            Estado civil
            <select
              value={answers.partner_marital_status}
              onChange={(event) =>
                updateAnswer("partner_marital_status", event.target.value)
              }
            >
              <option value="solteiro">Solteiro(a)</option>
              <option value="casado">Casado(a)</option>
              <option value="divorciado">Divorciado(a)</option>
            </select>
          </label>

          {conditionalFields.map((field) => (
            <TextField
              key={field}
              name={field}
              value={answers[field] || ""}
              missing={missingFields.includes(field)}
              onChange={updateAnswer}
            />
          ))}

          <div className="full">
            <button className="primary">Validar e concluir</button>
          </div>
        </form>

        <RulesPanel rules={rules} briefings={briefings} />
      </div>
    </>
  );
}
