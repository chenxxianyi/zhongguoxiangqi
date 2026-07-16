const DAY_MS = 24 * 60 * 60 * 1000

function parseDate(value: string): Date | null {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}

export function formatDate(value: string): string {
  const date = parseDate(value)
  if (!date) return value.slice(0, 10)

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function formatRelativeDate(value: string, now = new Date()): string {
  const date = parseDate(value)
  if (!date) return value.slice(0, 10)

  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const diffDays = Math.floor((startOfToday.getTime() - startOfDate.getTime()) / DAY_MS)

  if (diffDays === 0) return '今日'
  if (diffDays === 1) return '昨日'
  if (diffDays > 1 && diffDays < 30) return `${diffDays} 天前`
  return formatDate(value)
}
