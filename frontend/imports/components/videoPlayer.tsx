
import { css } from '@emotion/react'
import React, { SyntheticEvent, useState } from 'react'
import dynamic from 'next/dynamic'
import { Button } from '@mui/material'
import { Controls } from './controls'
import BaseReactPlayer, { BaseReactPlayerProps, OnProgressProps } from 'react-player/base'
// Fix for Hydration error
const ReactPlayer = dynamic(async () => await import('./reactPlayerWrapper'), { ssr: false })

const formatTime = (time: number) => {
  if (isNaN(time)) {
    return '00:00'
  }

  const date = new Date(time * 1000)
  const hours = date.getUTCHours()
  const minutes = date.getUTCMinutes()
  const seconds = date.getUTCSeconds().toString().padStart(2, '0')
  if (hours) {
    return `${hours}:${minutes.toString().padStart(2, '0')}`
  } else {
    return `${minutes}:${seconds}`
  }
}

const LoadFileButton = (props: { setFileUrl: (url: string) => void }) => {
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files?.length !== 1) {
      return
    }
    const file = e.target.files[0]
    const url = URL.createObjectURL(file)
    props.setFileUrl(url)
  }

  return (
    <Button
      component='label'
      variant='outlined'
      css={css`margin-right: 8px;`}
    >
      Select Video
      <input type='file' hidden onChange={handleFileSelect} />
    </Button>
  )
}

export const VideoPlayer = (props: { url?: string, videoName: string }) => {
  const [url, setUrl] = useState(props.url)

  const [playing, setPlaying] = useState(false)
  const [played, setPlayed] = useState(0)
  const [seeking, setSeeking] = useState(false)
  const [volume, setVolume] = useState(0.5)
  const [muted, setMuted] = useState(false)
  const [count, setCount] = useState(0)

  const controlRef = React.useRef<HTMLDivElement>(null)
  const videoPlayerRef = React.useRef<BaseReactPlayer<BaseReactPlayerProps>>(null)

  const currentTime = videoPlayerRef.current
    ? videoPlayerRef.current.getCurrentTime()
    : 0
  const duration = videoPlayerRef.current
    ? videoPlayerRef.current.getDuration()
    : 0

  const formatCurrentTime = formatTime(currentTime)
  const formatDuration = formatTime(duration)

  const progressHandler = (state: OnProgressProps) => {
    if (count > 3 && controlRef?.current) {
      controlRef.current.style.visibility = 'hidden'
    } else if (controlRef?.current?.style?.visibility === 'visible') {
      setCount(count + 1)
    }

    if (!seeking) {
      setPlayed(state.played)
    }
  }

  const playPauseHandler = () => {
    setPlaying(!playing)
  }

  const seekHandler = (_e: Event, value: number) => {
    const newPlayed = value / 100
    setPlayed(newPlayed)
    videoPlayerRef?.current?.seekTo(newPlayed)
  }

  const seekMouseUpHandler = (_e: SyntheticEvent<Element, Event>, value: number) => {
    setSeeking(false)
    videoPlayerRef?.current?.seekTo(value / 100)
  }

  const volumeChangeHandler = (_e: Event, value: number) => {
    setVolume(value / 100)
    setMuted(value === 0)
  }

  const volumeSeekUpHandler = (e: SyntheticEvent<Element, Event>, value: number) => {
    volumeChangeHandler(e.nativeEvent, value)
  }

  const muteHandler = () => {
    // TODO: verify volume is not 0
    setMuted(!muted)
  }

  const onSeekMouseDownHandler = () => {
    setSeeking(true)
  }

  const mouseMoveHandler = () => {
    if (!controlRef.current) {
      return
    }

    controlRef.current.style.visibility = 'visible'
  }

  return (
    <>
      {url
        ? (
          <div onMouseMove={mouseMoveHandler}>
            <ReactPlayer
              playerRef={videoPlayerRef}
              css={css`
                background-color: #000;
                margin: 10px 0;
              `}
              url={url}
              width={800}
              height={450}
              onProgress={progressHandler}
              playing={playing}
              volume={volume}
              muted={muted}
            />
            <Controls
              videoName={props.videoName}
              controlRef={controlRef}
              onPlayPause={playPauseHandler}
              playing={playing}
              played={played}
              onSeek={seekHandler}
              onMouseSeekUp={seekMouseUpHandler}
              onMouseSeekDown={onSeekMouseDownHandler}
              volume={volume}
              onVolumeChangeHandler={volumeChangeHandler}
              onVolumeSeekUp={volumeSeekUpHandler}
              mute={muted}
              onMute={muteHandler}
              duration={formatDuration}
              currentTime={formatCurrentTime}
            />
          </div>)
        : (
          <div css={css`
            background-color: #000;
            display: flex;
            justify-content: center;
            align-items: center;
            width: 800px;
            height: 450px;
            margin: 10px 0;
          `}
          >
            <LoadFileButton setFileUrl={setUrl} />
          </div>
          )}
    </>
  )
}
