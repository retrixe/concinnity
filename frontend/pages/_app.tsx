import React from 'react'
import Head from 'next/head'
import { AppProps } from 'next/app'
import createCache from '@emotion/cache'
import { RecoilRoot, useRecoilState } from 'recoil'
import { CacheProvider, EmotionCache } from '@emotion/react'
import { useMediaQuery, ThemeProvider, CssBaseline } from '@mui/material'
import { darkModeAtom } from '../imports/recoil-atoms'
import createTheme from '../imports/theme'

// Client-side cache, shared for the whole session of the user in the browser.
const clientSideEmotionCache = createCache({ key: 'css' })

function AppThemeProvider (props: React.PropsWithChildren<Record<string, unknown>>) {
  const [darkMode, setDarkMode] = useRecoilState(darkModeAtom)
  const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')
  const darkModeSetting = darkMode === true || (darkMode === undefined && prefersDarkMode)
  const theme = React.useMemo(() => createTheme(darkModeSetting), [darkModeSetting])

  React.useEffect(() => { // Read darkMode from localStorage.
    if (typeof localStorage === 'undefined') return
    const darkModePref = localStorage.getItem('darkMode')
    if (darkModePref === 'true') setDarkMode(true)
    else if (darkModePref === 'false') setDarkMode(false)
    else if (darkModePref === 'system') setDarkMode(undefined)
  }, [setDarkMode])

  return <ThemeProvider theme={theme}>{props.children}</ThemeProvider>
}

export default function MyApp (props: AppProps & { emotionCache?: EmotionCache }) {
  const { Component, emotionCache = clientSideEmotionCache, pageProps } = props

  return (
    <RecoilRoot>
      <CacheProvider value={emotionCache}>
        <Head>
          {/* Use minimum-scale=1 to enable GPU rasterization */}
          <meta
            name='viewport'
            content='user-scalable=0, initial-scale=1,
          minimum-scale=1, width=device-width, height=device-height'
          />
        </Head>
        <AppThemeProvider>
          {/* CssBaseline kickstart an elegant, consistent, and simple baseline to build upon. */}
          <CssBaseline />
          <React.StrictMode>
            <Component {...pageProps} />
          </React.StrictMode>
        </AppThemeProvider>
      </CacheProvider>
    </RecoilRoot>
  )
}
