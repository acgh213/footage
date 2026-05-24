import { writable } from 'svelte/store'

export const timePos = writable<number>(0)
export const paused = writable<boolean>(false)
export const speed = writable<number>(1)
export const mpvRunning = writable<boolean>(false)

// Poll interval reference so we can clear it
let pollInterval: ReturnType<typeof setInterval> | null = null

export function startPolling(getTimeFn: () => Promise<number>) {
  stopPolling()
  pollInterval = setInterval(async () => {
    try {
      const t = await getTimeFn()
      timePos.set(t)
    } catch {
      // mpv stopped; polling continues but won't update
    }
  }, 250)
}

export function stopPolling() {
  if (pollInterval !== null) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}
