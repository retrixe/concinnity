import React from 'react'
import { Typography } from '@mui/material'
import Title from '../imports/components/title'
import { AppDiv, TopBar } from '../imports/components/layout'

const IndexPage = () => {
  return (
    <>
      <Title title='Home - Concinnity' url='/' description='' />
      <TopBar />
      <AppDiv>
        <Typography variant='h5' align='center'>
          Watch video files together with others on the internet.
        </Typography>
        <Typography align='center'>Concinnity handles syncing up the video for you \o/</Typography>
      </AppDiv>
    </>
  )
}

export default IndexPage
