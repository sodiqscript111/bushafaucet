export default function LoadingState() {
  return (
    <div className="card">
      <div className="logo-row">
        <div className="logo-dot">💧</div>
        <span className="logo-name">Busha Faucet</span>
        <span className="badge">Sandbox</span>
      </div>
      <div className="centered">
        <div className="spinner"></div>
        <p className="screen-sub" style={{ marginTop: '1rem' }}>
          Sending tokens, please wait…
        </p>
      </div>
    </div>
  )
}
