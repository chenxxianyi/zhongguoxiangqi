import type { ApiErrorBody } from './contracts'

const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export class ApiError extends Error {
  constructor(public status: number, public body: ApiErrorBody) { super(body.message) }
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (typeof init.body === 'string' && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(`${baseUrl}${path}`, {
    ...init,
    headers,
    credentials: 'include',
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({ code: 'HTTP_ERROR', message: `请求失败（${response.status}）` })) as ApiErrorBody
    throw new ApiError(response.status, body)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export function createIdempotencyKey(prefix = 'web') {
  return `${prefix}-${crypto.randomUUID()}`
}
