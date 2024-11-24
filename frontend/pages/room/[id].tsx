import { Button, Typography } from '@mui/material'
import { useRouter } from 'next/router'
import React, { useEffect, useState } from 'react'
import config from '../../config.json'
import { AppDiv, FlexSpacer, TopBar } from '../../imports/components/helpers/layout'
import type { Room } from '../../imports/types'
import { VideoPlayer } from '../../imports/components/room/videoPlayer'
import LoginDialog from '../../imports/components/helpers/loginDialog'
import { useLoginStatus } from '../../imports/store'

const RoomPage = (): React.JSX.Element => {
  const { id, fileUrl } = useRouter().query

  const loginStatus = useLoginStatus(state => state.loginStatus)
  const [room, setRoom] = useState<Room | undefined>()
  const [error, setError] = useState('')
  const [loginDialog, setLoginDialog] = useState(false)

  useEffect(() => {
    if (typeof id !== 'string' || !id) return
    fetch(config.serverUrl + `/api/room/${id}`, {
      method: 'GET',
      headers: { Authentication: localStorage.getItem('token') ?? '' },
    })
      .then(async res => (await res.json()) as Room | { error: string })
      .then(json => {
        if ('error' in json) {
          setError(json.error)
        } else {
          setRoom(json)
        }
      })
      .catch(console.error)
  }, [id])

  useEffect(() => {
    setLoginDialog(!loginStatus)
  }, [loginStatus])

  // TODO: Add Home button to leave room
  return (
    <>
      <TopBar />
      <LoginDialog shown={loginDialog} handleClose={() => setLoginDialog(false)} />
      <AppDiv>
        <Typography variant='h4'>{room?.title}</Typography>
        <div>
          {error && <Typography color='error'>{error}</Typography>}
          {room?.extra && (
            <Typography>
              <b>{room?.extra && room.extra}</b>
              {' is being played.'}
            </Typography>
          )}
          {!error && room && (
            <VideoPlayer
              url={typeof fileUrl === 'string' ? fileUrl : undefined}
              videoName={room?.extra ?? 'Loading'}
            />
          )}
          {
            !error && (
              <Button variant='outlined'>Play New Video</Button>
            ) /* TODO: bring up something like the startWatchingDialog which changes the value of extra */
          }
        </div>
        <FlexSpacer />
      </AppDiv>
    </>
  )
}

export default RoomPage
