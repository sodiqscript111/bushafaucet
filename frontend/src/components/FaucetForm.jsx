import { useState, useEffect, useCallback } from 'react'
import { getConfig, submitClaim } from '../services/api'
import logo from '../assets/light.svg'

export default function FaucetForm({ onSubmitting, onSuccess, onError }) {
  const [config, setConfig] = useState(null)
  const [configLoading, setConfigLoading] = useState(true)
  const [configError, setConfigError] = useState(false)
  const [walletAddress, setWalletAddress] = useState('')
  const [blockchain, setBlockchain] = useState('')
  const [amount, setAmount] = useState('')
  const [fieldError, setFieldError] = useState('')

  const loadConfig = useCallback(() => {
    setConfigLoading(true)
    setConfigError(false)
    getConfig()
      .then((data) => {
        setConfig(data)
        setConfigLoading(false)
      })
      .catch(() => {
        setConfigError(true)
        setConfigLoading(false)
      })
  }, [])

  useEffect(() => {
    loadConfig()
  }, [loadConfig])

  const maxAmount = config?.max_amounts?.[blockchain] || ''

  const handleBlockchainChange = (e) => {
    const newNetwork = e.target.value
    setBlockchain(newNetwork)
    
    // Auto-fill the required amount to prevent minimum/maximum errors
    if (newNetwork && config?.max_amounts?.[newNetwork]) {
      setAmount(config.max_amounts[newNetwork])
    } else {
      setAmount('')
    }
  }

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
        <img src={logo} alt="Busha" className="logo-img" />
        <span className="logo-faucet">Faucet</span>
      </div>
      <h1 className="screen-title">Request test tokens</h1>
      <p className="screen-sub">
        Get sandbox crypto for testing your integration. Tokens are not real and
        have no monetary value.
      </p>

      {configLoading && (
        <div className="config-loading">
          <div className="spinner spinner-sm"></div>
          <span>Loading available networks…</span>
        </div>
      )}

      {configError && (
        <div className="error-banner">
          <span>⚠️</span>
          <span>
            Failed to load networks.{' '}
            <button type="button" className="retry-link" onClick={loadConfig}>
              Retry
            </button>
          </span>
        </div>
      )}

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
          onChange={handleBlockchainChange}
          disabled={configLoading || configError}
        >
          <option value="">
            {configLoading ? 'Loading networks…' : 'Select a network'}
          </option>
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
          readOnly={true}
        />
        {maxAmount && (
          <p className="field-hint">
            Amount is fixed at {maxAmount} {blockchain} by the sandbox API.
          </p>
        )}
      </div>

      <button
        type="submit"
        className="btn-primary"
        disabled={configLoading || configError}
      >
        Send test tokens
      </button>
    </form>
  )
}
