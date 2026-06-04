import logo from '../assets/light.svg'

export default function LoadingState() {
  return (
    <div className="card">
      <div className="logo-row">
        <img src={logo} alt="Busha" className="logo-img" />
        <span className="logo-faucet">Faucet</span>
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
