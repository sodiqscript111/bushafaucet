import { useState } from 'react'
import FaucetForm from './components/FaucetForm'
import LoadingState from './components/LoadingState'
import ErrorState from './components/ErrorState'
import SuccessState from './components/SuccessState'
import './App.css'

export default function App() {
  const [screen, setScreen] = useState('form')
  const [claim, setClaim] = useState(null)
  const [error, setError] = useState('')

  return (
    <div className="page">
      {screen === 'form' && (
        <FaucetForm
          onSubmitting={() => setScreen('loading')}
          onSuccess={(data) => {
            setClaim(data)
            setScreen('success')
          }}
          onError={(msg) => {
            setError(msg)
            setScreen('error')
          }}
        />
      )}
      {screen === 'loading' && <LoadingState />}
      {screen === 'error' && (
        <ErrorState message={error} onRetry={() => setScreen('form')} />
      )}
      {screen === 'success' && (
        <SuccessState claim={claim} onReset={() => setScreen('form')} />
      )}
    </div>
  )
}
