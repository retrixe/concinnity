import React from 'react'
import { useRouter } from 'next/router'
import { useRecoilValue } from 'recoil'
import { Button, Typography } from '@mui/material'
import Title from '../imports/components/title'
import { AppDiv, TopBar } from '../imports/components/layout'
import { loginStatusAtom } from '../imports/recoil-atoms'

const IndexPage = () => {
  const router = useRouter()
  const loginStatus = useRecoilValue(loginStatusAtom)
  React.useEffect(() => {
    if (!loginStatus) router.replace('/').catch(console.error)
  })
  return (
    <>
      <Title title='Home - Concinnity' url='/' description='' />
      <TopBar />
      <AppDiv>
        <Typography variant='h5' align='center'>
          Watch video files together with others on the internet.
        </Typography>
        <Typography align='center' gutterBottom>
          Concinnity handles syncing up the video for you \o/
        </Typography>
        <Typography align='center'>
          <Button variant='contained'>Start Watching</Button>
        </Typography>
      </AppDiv>
    </>
  )
}

export default IndexPage
