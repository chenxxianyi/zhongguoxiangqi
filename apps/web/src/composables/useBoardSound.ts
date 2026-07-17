import type { MoveRecord } from '@/api/contracts'

type ToneType = OscillatorType

export function useBoardSound() {
  let context: AudioContext | null = null

  function getContext() {
    if (typeof window === 'undefined' || !window.AudioContext) return null
    context ??= new window.AudioContext()
    if (context.state === 'suspended') void context.resume()
    return context
  }

  function tone(frequency: number, offset: number, duration: number, volume: number, type: ToneType = 'sine') {
    const audio = getContext()
    if (!audio) return
    const oscillator = audio.createOscillator()
    const gain = audio.createGain()
    const start = audio.currentTime + offset
    oscillator.type = type
    oscillator.frequency.setValueAtTime(frequency, start)
    gain.gain.setValueAtTime(0.0001, start)
    gain.gain.exponentialRampToValueAtTime(volume, start + 0.008)
    gain.gain.exponentialRampToValueAtTime(0.0001, start + duration)
    oscillator.connect(gain)
    gain.connect(audio.destination)
    oscillator.start(start)
    oscillator.stop(start + duration + 0.02)
  }

  function playMove(move: MoveRecord) {
    if (move.captured && move.givesCheck) {
      tone(196, 0, .1, .05, 'triangle')
      tone(131, .07, .15, .04, 'sine')
      tone(392, .16, .1, .038, 'triangle')
      tone(523, .25, .16, .035, 'triangle')
      return
    }
    if (move.givesCheck) {
      tone(392, 0, .11, .045, 'triangle')
      tone(523, .1, .18, .04, 'triangle')
      return
    }
    if (move.captured) {
      tone(196, 0, .1, .05, 'triangle')
      tone(131, .07, .16, .045, 'sine')
      return
    }
    tone(246, 0, .085, .035, 'triangle')
  }

  function playEnabled() {
    tone(330, 0, .07, .025, 'sine')
    tone(440, .07, .1, .025, 'sine')
  }

  function dispose() {
    if (context && context.state !== 'closed') void context.close()
    context = null
  }

  return { playMove, playEnabled, dispose }
}
