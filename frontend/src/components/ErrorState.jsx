import { useState } from 'react'
import logo from '../assets/light.svg'

function getFriendlyError(msg) {
  if (!msg) return 'An unexpected error occurred while processing your request. Please try again.'

  const lower = msg.toLowerCase()

  // Handle invalid address
  if (lower.includes('address is invalid') || lower.includes('addesss is invalid')) {
    return 'The wallet address you entered is invalid for the selected blockchain and network. Please double-check the address format and try again.'
  }

  // Handle amount issues (e.g. exceeds maximum)
  if (lower.includes('exceeds the maximum allowed')) {
    return msg
  }

  // Handle network / connection errors
  if (
    lower.includes('failed to fetch') ||
    lower.includes('network error') ||
    lower.includes('http request')
  ) {
    return 'We had trouble connecting to the server. Please check your internet connection and try again.'
  }

  // Handle general validation failure
  if (lower.includes('validation failed') || lower.includes('validationerror')) {
    const parts = msg.split('|')
    if (parts.length > 1) {
      return parts
        .slice(1)
        .map((p) => {
          const fieldParts = p.split(':')
          if (fieldParts.length >= 2) {
            const field = fieldParts[0].trim()
            const reason = fieldParts.slice(1).join(':').trim()
            const cleanField = field.replace('pay_out.', '').replace('_', ' ')
            const titleField = cleanField.charAt(0).toUpperCase() + cleanField.slice(1)
            return `${titleField}: ${reason}`
          }
          return p.trim()
        })
        .join('. ')
    }
    return 'Some of the details provided are invalid. Please check your inputs.'
  }

  // If the error message is too technical
  if (
    lower.includes('busha api error') ||
    lower.includes('internal server error') ||
    lower.includes('failed to create quote')
  ) {
    return 'The transaction failed to initialize on the test network. This might be due to temporary network issues or rate limiting.'
  }

  return msg
}

export default function ErrorState({ message, onRetry }) {
  const [showDetails, setShowDetails] = useState(false)
  const friendlyMessage = getFriendlyError(message)

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
        <p className="screen-sub" style={{ marginBottom: '1.5rem', color: '#e53e3e', fontWeight: 500 }}>
          {friendlyMessage}
        </p>

        {message && message !== friendlyMessage && (
          <div style={{ width: '100%', marginBottom: '1.5rem' }}>
            <button
              type="button"
              className="details-toggle"
              onClick={() => setShowDetails(!showDetails)}
            >
              {showDetails ? 'Hide' : 'Show'} technical details {showDetails ? '▲' : '▼'}
            </button>
            {showDetails && <pre className="details-pre">{message}</pre>}
          </div>
        )}

        <button className="btn-primary" onClick={onRetry}>
          Try again
        </button>
      </div>
    </div>
  )
}
