import React from 'react'
import { useRouter } from 'next/router'
import { useRecoilValue } from 'recoil'
import { Button, Typography } from '@mui/material'
import Title from '../imports/components/title'
import { AppDiv, TopBar } from '../imports/components/layout'
import { loginStatusAtom } from '../imports/recoil-atoms'
import StartWatchingDialog from '../imports/components/startWatchingDialog'

const IndexPage = () => {
  const [startWatchingShown, setStartWatchingShown] = React.useState(false)

  const router = useRouter()
  const loginStatus = useRecoilValue(loginStatusAtom)
  React.useEffect(() => {
    if (!loginStatus) router.replace('/').catch(console.error)
  })

  return (
    <>
      <StartWatchingDialog shown={startWatchingShown} handleClose={() => setStartWatchingShown(false)} />
      <Title
        title='Home - Concinnity' url='/home'
        description='Concinnity - Watch video files together with others on the internet.'
      />
      <TopBar />
      <AppDiv>
        <Typography variant='h5' align='center'>
          Watch video files together with others on the internet.
        </Typography>
        <Typography align='center' gutterBottom>
          Concinnity handles syncing up the video for you \o/
        </Typography>
        <Typography align='center'>
          <Button variant='contained' onClick={() => setStartWatchingShown(true)}>Start Watching</Button>
        </Typography>
      </AppDiv>
    </>
  )
}

export default IndexPage
