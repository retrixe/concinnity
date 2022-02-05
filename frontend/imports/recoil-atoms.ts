import { atom } from 'recoil'

export const darkModeAtom = atom<boolean | undefined>({
  key: 'darkMode',
  default: true
})
