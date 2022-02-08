import React from 'react'
import styled from '@emotion/styled'
import { useRecoilState } from 'recoil'
import { AppBar, IconButton, Toolbar, Typography } from '@mui/material'
import SettingsBrightnessOutlined from '@mui/icons-material/SettingsBrightnessOutlined'
import LightModeOutlined from '@mui/icons-material/LightModeOutlined'
import DarkModeOutlined from '@mui/icons-material/DarkModeOutlined'
import { darkModeAtom } from '../recoil-atoms'

const TopBarCenteredContent = styled.div({})

export const FlexSpacer = styled.div({ flex: 1 })

export const TopBar = (props: { variant?: 'dense' }) => {
  const [darkMode, setDarkMode] = useRecoilState(darkModeAtom) // System then Dark then Light
  const themeToggle = () => setDarkMode(state => state === false ? undefined : state !== true)
  return (
    <AppBar position='static' enableColorOnDark elevation={1}>
      <TopBarCenteredContent>
        <Toolbar variant={props.variant}>
          <Typography variant='h6'>Concinnity</Typography>
          <FlexSpacer />
          <IconButton color='inherit' onClick={themeToggle}>
            {darkMode === true
              ? <DarkModeOutlined />
              : (darkMode === false ? <LightModeOutlined /> : <SettingsBrightnessOutlined />)}
          </IconButton>
        </Toolbar>
      </TopBarCenteredContent>
    </AppBar>
  )
}

export const AppDiv = styled.div({ margin: '16px' })
