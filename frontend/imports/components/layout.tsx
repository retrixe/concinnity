import React from 'react'
import styled from '@emotion/styled'
import { useRouter } from 'next/router'
import { useRecoilState, useRecoilValue } from 'recoil'
import { AppBar, IconButton, Toolbar, Tooltip, Typography } from '@mui/material'
import SettingsBrightnessOutlined from '@mui/icons-material/SettingsBrightnessOutlined'
import LightModeOutlined from '@mui/icons-material/LightModeOutlined'
import DarkModeOutlined from '@mui/icons-material/DarkModeOutlined'
import Logout from '@mui/icons-material/Logout'
import Login from '@mui/icons-material/Login'
import config from '../../config.json'
import { darkModeAtom, loginStatusAtom } from '../recoil-atoms'

const TopBarCenteredContent = styled.div({})

export const FlexSpacer = styled.div({ flex: 1 })

export const TopBar = (props: { variant?: 'dense' }) => {
  const router = useRouter()
  const loginStatus = useRecoilValue(loginStatusAtom)
  const [darkMode, setDarkMode] = useRecoilState(darkModeAtom) // System then Dark then Light

  const themeToggle = () => setDarkMode(state => state === false ? undefined : state !== true)
  const handleLogin = () => {
    if (loginStatus) {
      fetch(config.serverUrl + '/logout').catch(console.error)
      router.push('/').catch(console.error)
    }
  }

  return (
    <AppBar position='static' enableColorOnDark elevation={1}>
      <TopBarCenteredContent>
        <Toolbar variant={props.variant}>
          <Typography variant='h6'>Concinnity</Typography>
          <FlexSpacer />
          <IconButton color='inherit' onClick={handleLogin}>
            <Tooltip title={loginStatus ? 'Logout' : 'Login'}>
              {loginStatus ? <Logout /> : <Login />}
            </Tooltip>
          </IconButton>
          <IconButton color='inherit' onClick={themeToggle}>
            <Tooltip title='Theme'>
              {darkMode === true
                ? <DarkModeOutlined />
                : (darkMode === false ? <LightModeOutlined /> : <SettingsBrightnessOutlined />)}
            </Tooltip>
          </IconButton>
        </Toolbar>
      </TopBarCenteredContent>
    </AppBar>
  )
}

export const AppDiv = styled.div({ margin: '16px' })