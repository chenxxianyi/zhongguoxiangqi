import { apiRequest, createIdempotencyKey } from './client'
import type { LegalMovesResponse, MatchAIMode, MatchSnapshot } from './contracts'

export interface CreateMatchRequest {
  playerColor: 'red' | 'black'
  difficulty: number
  aiMode: MatchAIMode
  allowUndo: boolean
}

export interface MatchVersionRequest {
  expectedMatchVersion: number
}

export interface SubmitMoveRequest extends MatchVersionRequest {
  move: string
}

export interface DrawOfferResponse {
  accepted: boolean
  match: MatchSnapshot
}

export async function listMatches(): Promise<MatchSnapshot[]> {
  const response = await apiRequest<{ items: MatchSnapshot[] }>('/matches')
  return response.items
}

export function createMatch(request: CreateMatchRequest): Promise<MatchSnapshot> {
  return apiRequest<MatchSnapshot>('/matches', {
    method: 'POST',
    headers: { 'Idempotency-Key': createIdempotencyKey('match-create') },
    body: JSON.stringify(request),
  })
}

export function getMatch(id: string): Promise<MatchSnapshot> {
  return apiRequest<MatchSnapshot>(`/matches/${encodeURIComponent(id)}`)
}

export function getLegalMoves(id: string, from?: string): Promise<LegalMovesResponse> {
  const query = from ? `?from=${encodeURIComponent(from)}` : ''
  return apiRequest<LegalMovesResponse>(
    `/matches/${encodeURIComponent(id)}/legal-moves${query}`,
  )
}

export function submitMatchMove(
  id: string,
  request: SubmitMoveRequest,
): Promise<MatchSnapshot> {
  return apiRequest<MatchSnapshot>(`/matches/${encodeURIComponent(id)}/moves`, {
    method: 'POST',
    headers: { 'Idempotency-Key': createIdempotencyKey('match-move') },
    body: JSON.stringify(request),
  })
}

export function undoMatch(id: string, request: MatchVersionRequest): Promise<MatchSnapshot> {
  return apiRequest<MatchSnapshot>(`/matches/${encodeURIComponent(id)}/undo`, {
    method: 'POST',
    headers: { 'Idempotency-Key': createIdempotencyKey('match-undo') },
    body: JSON.stringify(request),
  })
}

export function resignMatch(id: string, request: MatchVersionRequest): Promise<MatchSnapshot> {
  return apiRequest<MatchSnapshot>(`/matches/${encodeURIComponent(id)}/resign`, {
    method: 'POST',
    headers: { 'Idempotency-Key': createIdempotencyKey('match-resign') },
    body: JSON.stringify(request),
  })
}

export function offerMatchDraw(
  id: string,
  request: MatchVersionRequest,
): Promise<DrawOfferResponse> {
  return apiRequest<DrawOfferResponse>(`/matches/${encodeURIComponent(id)}/draw-offers`, {
    method: 'POST',
    headers: { 'Idempotency-Key': createIdempotencyKey('match-draw') },
    body: JSON.stringify(request),
  })
}
