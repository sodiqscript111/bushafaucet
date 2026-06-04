const BASE_URL = 'http://localhost:8085/api/v1'

export async function getConfig() {
  const res = await fetch(`${BASE_URL}/config`)
  if (!res.ok) throw new Error('Failed to load config')
  return res.json()
}

export async function submitClaim(data) {
  const res = await fetch(`${BASE_URL}/faucet`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  const json = await res.json()
  if (!res.ok) throw new Error(json.error || 'Something went wrong')
  return json
}

export async function getClaim(id) {
  const res = await fetch(`${BASE_URL}/claims/${id}`)
  if (!res.ok) throw new Error('Failed to fetch claim')
  return res.json()
}
