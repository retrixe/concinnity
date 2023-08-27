import { Typography } from '@mui/material'
import { useRouter } from 'next/router'
import React, { useState } from 'react'
import { useRecoilValue } from 'recoil'
import config from '../../config.json'
import { loginStatusAtom } from '../../imports/recoil-atoms'
import { AppDiv, FlexSpacer, TopBar } from '../../imports/components/helpers/layout'
import { type Room } from '../../imports/types'
import { VideoPlayer } from '../../imports/components/room/videoPlayer'
import LoginDialog from '../../imports/components/helpers/loginDialog'

const RoomPage = (): JSX.Element => {
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
        <div>
          {error && <Typography color='error'>{error}</Typography>}
          <VideoPlayer
            url={typeof fileUrl === 'string' ? fileUrl : undefined}
            videoName={room?.extra ?? 'Loading'}
          />
          {room?.extra &&
            <Typography>
              <b>{room?.extra && room.extra}</b>
              {' is being played.'}
            </Typography>}
        </div>
        <FlexSpacer />
      </AppDiv>
    </>
  )
}

export default RoomPage
