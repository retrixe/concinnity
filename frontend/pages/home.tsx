import React from 'react'
import { useRouter } from 'next/router'
import { Button, Typography } from '@mui/material'
import Title from '../imports/components/helpers/title'
import { AppDiv, TopBar } from '../imports/components/helpers/layout'
import StartWatchingDialog from '../imports/components/home/startWatchingDialog'
import { useLoginStatus } from '../imports/store'

const IndexPage = (): React.JSX.Element => {
  const [startWatchingShown, setStartWatchingShown] = React.useState(false)

  const router = useRouter()
  const loginStatus = useLoginStatus(state => state.loginStatus)
  React.useEffect(() => {
    if (loginStatus === false) router.replace('/').catch(console.error)
  })

  return (
    <>
      <StartWatchingDialog
        shown={startWatchingShown}
        handleClose={() => setStartWatchingShown(false)}
      />
      <Title
        title='Home - Concinnity'
        url='/home'
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
          <Button variant='contained' onClick={() => setStartWatchingShown(true)}>
            Start Watching
          </Button>
        </Typography>
      </AppDiv>
    </>
  )
}

export default IndexPage
