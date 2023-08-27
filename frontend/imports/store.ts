import { useMediaQuery } from '@mui/material'
import { create } from 'zustand'
import config from '../config.json'

export const useDarkMode = create<{
  darkMode: boolean | undefined
  setDarkMode: (newValue: boolean | undefined) => void
  loadDarkModeSetting: () => void
  useDarkModeValue: () => boolean
}>((set, get) => ({ // Value is initialised in _app.tsx
      darkMode: true,
      useDarkModeValue: () => {
        const darkModeSetting = get().darkMode
        const systemDarkModeValue = useMediaQuery('(prefers-color-scheme: dark)')
        return darkModeSetting ?? systemDarkModeValue
      },
      loadDarkModeSetting: () => {
        if (typeof localStorage === 'undefined') return
        const darkMode = localStorage.getItem('darkMode')
        if (darkMode === 'true') set({ darkMode: true })
        else if (darkMode === 'false') set({ darkMode: false })
        else if (darkMode === 'system') set({ darkMode: undefined })
      },
      setDarkMode: (newValue: boolean | undefined) => {
        if (typeof localStorage === 'undefined') return
        // else if (newValue === true) localStorage.setItem('darkMode', 'true')
        else if (newValue === true) localStorage.removeItem('darkMode')
        else if (newValue === false) localStorage.setItem('darkMode', 'false')
        else if (newValue === undefined) localStorage.setItem('darkMode', 'system')
        set({ darkMode: newValue })
      }
    }))

export const useLoginStatus = create<{
  loginStatus: string | false
  setLoginStatus: (newValue: string | false) => void
  loadLoginStatus: () => void
}>((set) => ({
      loginStatus: false,
      setLoginStatus: (newValue: string | false) => set({ loginStatus: newValue }),
      loadLoginStatus: (): void => {
        const token = localStorage.getItem('token')
        if (!token) return
        // If loading, set self to empty
        set({ loginStatus: '' })
        fetch(config.serverUrl, { headers: { Authentication: token } })
          .then(async res => await res.json())
          .then(res => set(({ loginStatus: res.username })))
          .catch(console.error)
      }
    }))
