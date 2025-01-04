export const stringifyDuration = (duration: number): string => {
  const seconds = Math.floor(duration % 60)
  const minutes = Math.floor((duration / 60) % 60)
  const hours = Math.floor(duration / 3600)

  const hoursString = hours > 0 ? `${hours}:` : ''
  const minutesString = `${minutes < 10 ? '0' : ''}${minutes}:`
  const secondsString = `${seconds < 10 ? '0' : ''}${seconds}`

  return hoursString + minutesString + secondsString
}
