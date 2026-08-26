export function Brand({ compact = false }) {
  return (
    <div className={compact ? "brand compact" : "brand"}>
      <div className="mark">S</div>
      <div>
        <strong>Sheep</strong>
        <span>CONTABIL</span>
      </div>
    </div>
  );
}
