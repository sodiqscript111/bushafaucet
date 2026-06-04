import logo from '../assets/light.svg'

export default function ErrorState({ message, onRetry }) {
  return (
    <div className="card">
      <div className="logo-row">
        <img src={logo} alt="Busha" className="logo-img" />
        <span className="logo-faucet">Faucet</span>
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
