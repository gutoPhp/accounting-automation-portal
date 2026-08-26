export function Toast({ value }) {
  if (!value) {
    return null;
  }

  return <div className={`toast ${value.type || ""}`}>{value.text}</div>;
}

export function EmptyState({ children }) {
  return <div className="empty">{children}</div>;
}
