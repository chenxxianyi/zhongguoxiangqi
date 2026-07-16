import { describe, expect, it } from 'vitest'
import { formatDate, formatRelativeDate } from './date'

describe('date helpers', () => {
  const now = new Date(2026, 6, 16, 12)

  it('formats recent calendar dates relatively', () => {
    expect(formatRelativeDate('2026-07-16T01:00:00+08:00', now)).toBe('今日')
    expect(formatRelativeDate('2026-07-15T01:00:00+08:00', now)).toBe('昨日')
    expect(formatRelativeDate('2026-07-10T01:00:00+08:00', now)).toBe('6 天前')
  })

  it('falls back to an absolute date for older values', () => {
    expect(formatRelativeDate('2026-06-01T01:00:00+08:00', now)).toBe('2026-06-01')
    expect(formatDate('invalid-date')).toBe('invalid-da')
  })
})
