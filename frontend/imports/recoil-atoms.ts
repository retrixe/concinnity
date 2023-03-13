import { atom } from 'recoil'
import config from '../config.json'

export const darkModeAtom = atom<boolean | undefined>({
  key: 'darkMode',
  default: true,
  effects: [({ onSet }) => { // setSelf is handled in React to ensure proper SSR hydration.
    if (typeof localStorage === 'undefined') return
    onSet((newValue, _, isReset) => {
      if (isReset) localStorage.removeItem('darkMode')
      else if (newValue === true) localStorage.setItem('darkMode', 'true')
      else if (newValue === false) localStorage.setItem('darkMode', 'false')
      else if (newValue === undefined) localStorage.setItem('darkMode', 'system')
    })
  }]
})

export const loginStatusAtom = atom<false | string>({
  key: 'loginStatus',
  default: false,
  effects: [({ setSelf }) => {
    if (typeof localStorage === 'undefined') return
    const loadAtomState = () => {
      const token = localStorage.getItem('token')
      if (!token) return
      // If loading, set self to empty
      setSelf('')
      fetch(config.serverUrl, { headers: { Authentication: token } })
        .then(async res => await res.json())
        .then(res => setSelf(res.username))
        .catch(console.error)
    }
    loadAtomState()
    setInterval(loadAtomState, 5000)
  }]
})
