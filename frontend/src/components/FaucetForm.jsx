import { useState, useEffect } from 'react'
import { getConfig, submitClaim } from '../services/api'

export default function FaucetForm({ onSubmitting, onSuccess, onError }) {
  const [config, setConfig] = useState(null)
  const [walletAddress, setWalletAddress] = useState('')
  const [blockchain, setBlockchain] = useState('')
  const [amount, setAmount] = useState('')
  const [fieldError, setFieldError] = useState('')

  useEffect(() => {
    getConfig()
      .then(setConfig)
      .catch(() =>
        setConfig({
          blockchains: ['BTC', 'ETH', 'USDT', 'USDC', 'BNB'],
          max_amounts: {
            BTC: '0.0001',
            ETH: '0.5',
            USDT: '5',
            USDC: '5',
            BNB: '0.1',
          },
        }),
      )
  }, [])

  const maxAmount = config?.max_amounts?.[blockchain] || ''

  const handleSubmit = async (e) => {
    e.preventDefault()
    setFieldError('')

    if (!walletAddress) return setFieldError('Wallet address is required')
    if (!blockchain) return setFieldError('Please select a network')
    if (!amount || Number(amount) <= 0)
      return setFieldError('Enter a valid amount')
    if (maxAmount && Number(amount) > Number(maxAmount)) {
      return setFieldError(`Max amount for ${blockchain} is ${maxAmount}`)
    }

    onSubmitting()
    try {
      const result = await submitClaim({
        wallet_address: walletAddress,
        blockchain,
        amount: Number(amount),
      })
      onSuccess(result.data)
    } catch (err) {
      onError(err.message)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="card">
      <div className="logo-row">
        <div className="logo-dot">💧</div>
        <span className="logo-name">Busha Faucet</span>
        <span className="badge">Sandbox</span>
      </div>
      <h1 className="screen-title">Request test tokens</h1>
      <p className="screen-sub">
        Get sandbox crypto for testing your integration. Tokens are not real and
        have no monetary value.
      </p>

      {fieldError && (
        <div className="error-banner">
          <span>⚠️</span>
          <span>{fieldError}</span>
        </div>
      )}

      <div className="field">
        <label>Wallet address</label>
        <input
          type="text"
          placeholder="Enter your sandbox wallet address"
          value={walletAddress}
          onChange={(e) => setWalletAddress(e.target.value)}
        />
      </div>

      <div className="field">
        <label>Network</label>
        <select
          value={blockchain}
          onChange={(e) => setBlockchain(e.target.value)}
        >
          <option value="">Select a network</option>
          {config?.blockchains?.map((b) => (
            <option key={b} value={b}>
              {b}
            </option>
          ))}
        </select>
      </div>

      <div className="field">
        <label>Amount</label>
        <input
          type="number"
          placeholder="0.00"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          min="0"
          step="any"
        />
        {maxAmount && (
          <p className="field-hint">
            Max {maxAmount} {blockchain}
          </p>
        )}
      </div>

      <button type="submit" className="btn-primary">
        Send test tokens
      </button>
    </form>
  )
}
