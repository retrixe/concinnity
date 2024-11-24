import { type Theme, createTheme as createMuiTheme } from '@mui/material'

const createTheme = (darkMode: boolean): Theme =>
  createMuiTheme({
    palette: {
      mode: darkMode ? 'dark' : 'light',
    },
  })

export default createTheme
