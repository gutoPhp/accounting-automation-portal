export function PageHeader({ code, title, subtitle, children }) {
  return (
    <header className="page-header">
      <div>
        {code && <span className="module-code">{code}</span>}
        <h1>{title}</h1>
        <p>{subtitle}</p>
      </div>
      {children}
    </header>
  );
}
