export function Metric({ value, label, tone = "" }) {
  return (
    <article className={`metric ${tone}`}>
      <strong>{String(value).padStart(2, "0")}</strong>
      <span>{label}</span>
    </article>
  );
}
