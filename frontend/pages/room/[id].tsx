import { Typography } from '@mui/material'
import { useRouter } from 'next/router'
import React, { useState } from 'react'
import { useRecoilValue } from 'recoil'
import config from '../../config.json'
import { loginStatusAtom } from '../../imports/recoil-atoms'
import { AppDiv, TopBar } from '../../imports/components/layout'
import { Room } from '../../imports/types'
import { VideoPlayer } from '../../imports/components/videoPlayer'
import LoginDialog from '../../imports/components/loginDialog'

const RoomPage = () => {
  const { id, fileUrl } = useRouter().query

  const loginStatus = useRecoilValue(loginStatusAtom)
  const [room, setRoom] = useState<Room | undefined>()
  const [error, setError] = useState('')
  const [loginDialog, setLoginDialog] = useState(false)

  React.useEffect(() => {
    if (typeof id !== 'string' || !id) return
    fetch(config.serverUrl + `/api/room/${id}`, {
      method: 'GET',
      headers: { Authentication: localStorage.getItem('token') ?? '' }
    }).then(async res => await res.json()).then(json => {
      if (json.error) {
        setError(json.error)
      } else {
        setRoom(json)
      }
    }).catch(console.error)
  }, [id])

  React.useEffect(() => {
    if (!loginStatus && loginStatus !== '') {
      setLoginDialog(true)
    } else {
      setLoginDialog(false)
    }
  }, [loginStatus])

  return (
    <>
      <TopBar />
      <LoginDialog shown={loginDialog} handleClose={() => setLoginDialog(false)} />
      <AppDiv>
        <Typography variant='h4'>{room?.title}</Typography>
        {error && <Typography color='error'>{error}</Typography>}
        <VideoPlayer url={typeof fileUrl === 'string' ? fileUrl : undefined} />
        <Typography>
          <b>{room?.extra && room.extra}</b>
          {' is being played.'}
        </Typography>
      </AppDiv>
    </>
  )
}

export default RoomPage
