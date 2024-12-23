import React, { useState } from 'react'
import styled from '@emotion/styled'
import { AppBar, IconButton, Toolbar, Tooltip, Typography } from '@mui/material'
import SettingsBrightnessOutlined from '@mui/icons-material/SettingsBrightnessOutlined'
import LightModeOutlined from '@mui/icons-material/LightModeOutlined'
import DarkModeOutlined from '@mui/icons-material/DarkModeOutlined'
import Logout from '@mui/icons-material/Logout'
import Login from '@mui/icons-material/Login'
import config from '../../../config.json'
import LoginDialog from './loginDialog'
import { useRouter } from 'next/router'
import { useDarkMode, useLoginStatus } from '../../store'

const TopBarCenteredContent = styled.div({})

export const FlexSpacer = styled.div({ flex: 1 })

export const TopBar = (props: { variant?: 'dense' }): React.JSX.Element => {
  const { darkMode, setDarkMode } = useDarkMode() // System then Dark then Light
  const { loginStatus, setLoginStatus } = useLoginStatus()
  const [loginDialog, setLoginDialog] = useState(false)
  const router = useRouter()

  const themeToggle = (): void => setDarkMode(darkMode === false ? undefined : darkMode !== true)
  const handleLogin = (): void => {
    const token = localStorage.getItem('token')
    if (loginStatus && token) {
      fetch(config.serverUrl + '/api/logout', {
        method: 'POST',
        headers: { Authentication: token },
      })
        .then(() => localStorage.removeItem('token'))
        .then(() => setLoginStatus(false))
        .then(async () => await router.replace('/'))
        .catch(console.error)
    } else setLoginDialog(true)
  }

  return (
    <>
      <LoginDialog shown={loginDialog} handleClose={() => setLoginDialog(false)} />
      <AppBar position='static' enableColorOnDark elevation={1}>
        <TopBarCenteredContent>
          <Toolbar variant={props.variant}>
            <Typography variant='h6'>Concinnity</Typography>
            <FlexSpacer />
            <IconButton color='inherit' onClick={handleLogin}>
              <Tooltip title={loginStatus ? 'Logout' : 'Login'}>
                {loginStatus !== false ? <Logout /> : <Login />}
              </Tooltip>
            </IconButton>
            <IconButton color='inherit' onClick={themeToggle}>
              <Tooltip title='Theme'>
                {darkMode === true ? (
                  <DarkModeOutlined />
                ) : darkMode === false ? (
                  <LightModeOutlined />
                ) : (
                  <SettingsBrightnessOutlined />
                )}
              </Tooltip>
            </IconButton>
          </Toolbar>
        </TopBarCenteredContent>
      </AppBar>
    </>
  )
}

export const AppDiv = styled.div({ margin: '16px' })
