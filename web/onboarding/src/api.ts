import type { Captcha, User, UserEvent } from './types.ts'

const BASE = '/api'
const CAPTCHA_BASE = '/captcha'

async function handle<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }
  return res.json()
}

export async function register(name: string, email: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, email }),
  }))
}

export async function verifyEmail(id: string, token: string): Promise<User> {
   const url = `${BASE}/users/${id}/verify-email`
  console.log("verifyEmail id:", id)
  console.log("verifyEmail url:", url)
  return handle(await fetch(`${BASE}/users/${id}/verify-email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  }))
}

export async function getUser(id: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/${id}`))
}

export async function listActivated(): Promise<User[]> {
  return handle(await fetch(`${BASE}/users/activated`))
}

export async function updateContact(id: string, name: string, email: string, bio: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, email, bio }),
  }))
}

export async function getUserHistory(id: string): Promise<UserEvent[]> {
  return handle(await fetch(`${BASE}/users/${id}/history`))
}

export async function completeProfile(id: string, bio: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/${id}/complete-profile`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bio }),
  }))
}

export async function issueCaptcha(): Promise<Captcha> {
  return handle(await fetch(CAPTCHA_BASE, { method: 'POST' }))
}

export async function verifyCaptcha(id: string, answer: number): Promise<void> {
  await handle<{ verified: boolean }>(await fetch(`${CAPTCHA_BASE}/${id}/verify`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ answer }),
  }))
}
