export default function ErrorState({ message, onRetry }) {
  return (
    <div className="card">
      <div className="logo-row">
        <div className="logo-dot">💧</div>
        <span className="logo-name">Busha Faucet</span>
        <span className="badge">Sandbox</span>
      </div>
      <div className="centered">
        <div className="error-icon">✕</div>
        <h2 className="screen-title" style={{ marginTop: '1rem' }}>
          Request failed
        </h2>
        <p className="screen-sub">{message}</p>
        <button className="btn-primary" onClick={onRetry}>
          Try again
        </button>
      </div>
    </div>
  )
}
