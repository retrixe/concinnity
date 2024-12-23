import React, { useEffect } from 'react'
import Head from 'next/head'
import type { AppProps } from 'next/app'
import createCache from '@emotion/cache'
import { CacheProvider, type EmotionCache } from '@emotion/react'
import { ThemeProvider, CssBaseline } from '@mui/material'
import { useDarkMode, useLoginStatus } from '../imports/store'
import createTheme from '../imports/theme'

// Client-side cache, shared for the whole session of the user in the browser.
const clientSideEmotionCache = createCache({ key: 'css' })

function AppThemeProvider(
  props: React.PropsWithChildren<Record<string, unknown>>,
): React.JSX.Element {
  const { loadLoginStatus } = useLoginStatus()
  const { useDarkModeValue, loadDarkModeSetting } = useDarkMode()
  const darkMode = useDarkModeValue()
  const theme = React.useMemo(() => createTheme(darkMode), [darkMode])

  useEffect(loadDarkModeSetting, [loadDarkModeSetting]) // Read darkMode from localStorage.
  useEffect(() => {
    // Load login status.
    loadLoginStatus()
    const interval = setInterval(loadLoginStatus, 5000)
    return () => clearInterval(interval)
  }, [loadLoginStatus])

  return <ThemeProvider theme={theme}>{props.children}</ThemeProvider>
}

export default function MyApp(
  props: AppProps & { emotionCache?: EmotionCache },
): React.JSX.Element {
  const { Component, emotionCache = clientSideEmotionCache, pageProps } = props

  return (
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
  )
}
