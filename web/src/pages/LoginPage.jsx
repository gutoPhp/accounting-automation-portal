import { useState } from "react";

import { Brand } from "../components/ui/Brand";
import { apiRequest } from "../lib/api";

const initialCredentials = {
  email: "admin@sheepcontabil.com",
  password: "Sheep@2026",
};

export function LoginPage({ onLogin }) {
  const [credentials, setCredentials] = useState(initialCredentials);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function updateField(field, value) {
    setCredentials((current) => ({ ...current, [field]: value }));
  }

  async function handleSubmit(event) {
    event.preventDefault();
    setSubmitting(true);
    setError("");

    try {
      const user = await apiRequest("/api/login", {
        method: "POST",
        body: JSON.stringify(credentials),
      });
      onLogin(user);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="login-page">
      <section className="login-story">
        <Brand />
        <div>
          <span className="kicker">AUTOMAÇÃO COM CONTROLE</span>
          <h1>
            Menos operação manual.
            <br />
            Mais clareza para decidir.
          </h1>
          <p>
            Processos contábeis centralizados, rastreáveis e preparados para
            falhar com dignidade.
          </p>
        </div>
        <small>Ambiente demonstrativo · dados inteiramente sintéticos</small>
      </section>

      <section className="login-panel">
        <form className="login-card" onSubmit={handleSubmit}>
          <div>
            <span className="kicker">PORTAL OPERACIONAL</span>
            <h2>Boas-vindas</h2>
            <p>Entre com seu acesso para continuar.</p>
          </div>

          <label>
            E-mail
            <input
              value={credentials.email}
              onChange={(event) => updateField("email", event.target.value)}
              type="email"
              required
            />
          </label>

          <label>
            Senha
            <input
              value={credentials.password}
              onChange={(event) => updateField("password", event.target.value)}
              type="password"
              required
            />
          </label>

          {error && <div className="alert error">{error}</div>}

          <button className="primary" disabled={submitting}>
            {submitting ? "Entrando…" : "Entrar no portal"}
          </button>

          <p className="hint">
            Operador: operador@sheepcontabil.com · mesma senha
          </p>
        </form>
      </section>
    </main>
  );
}
