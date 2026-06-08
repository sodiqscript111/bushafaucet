import logo from '../assets/light.svg'

export default function SuccessState({ claim, onReset }) {
  return (
    <div className="card">
      <div className="logo-row">
        <img src={logo} alt="Busha" className="logo-img" />
        <span className="logo-faucet">Faucet</span>
      </div>
      <div className="centered">
        <div className="success-icon">✓</div>
        <h2 className="screen-title" style={{ marginTop: '1rem' }}>
          Tokens sent successfully
        </h2>
        <p className="screen-sub">
          Your test tokens are on the way. They may take a few seconds to appear
          in your sandbox wallet.
        </p>
        <div className="receipt">
          <div className="receipt-row">
            <span>Asset</span>
            <span>{claim?.blockchain}</span>
          </div>
          <div className="receipt-row">
            <span>Network</span>
            <span>{claim?.network}</span>
          </div>
          <div className="receipt-row">
            <span>Amount</span>
            <span>
              {claim?.amount} {claim?.blockchain}
            </span>
          </div>
          <div className="receipt-row">
            <span>Address</span>
            <span>
              {claim?.wallet_address?.slice(0, 10)}...
              {claim?.wallet_address?.slice(-4)}
            </span>
          </div>
          <div className="receipt-row">
            <span>Status</span>
            <span>{claim?.status}</span>
          </div>
        </div>
        <button className="btn-primary" onClick={onReset}>
          Request again
        </button>
      </div>
    </div>
  )
}
