import { createTheme as createMuiTheme } from '@mui/material'

const createTheme = (darkMode: boolean) => createMuiTheme({
  palette: {
    mode: darkMode ? 'dark' : 'light'
  }
})

export default createTheme
