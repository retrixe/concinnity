import { css } from '@emotion/react'
import React, { type SyntheticEvent, useState } from 'react'
import dynamic from 'next/dynamic'
import { Button } from '@mui/material'
import { Controls } from './controls'
import { type BaseReactPlayerProps, type OnProgressProps } from 'react-player/base'
import type BaseReactPlayer from 'react-player/base'
// Fix for Hydration error
const ReactPlayer = dynamic(async () => await import('./reactPlayerWrapper'), { ssr: false })

const formatTime = (time: number): string => {
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

const LoadFileButton = (props: { setFileUrl: (url: string) => void, fileName: string }): JSX.Element => {
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>): void => {
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
      Select Video: {props.fileName}
      <input type='file' hidden onChange={handleFileSelect} />
    </Button>
  )
}

export const VideoPlayer = (props: { url?: string, videoName: string }): JSX.Element => {
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

  const progressHandler = (state: OnProgressProps): void => {
    if (count > 3 && controlRef?.current) {
      controlRef.current.style.visibility = 'hidden'
    } else if (controlRef?.current?.style?.visibility === 'visible') {
      setCount(count + 1)
    }

    if (!seeking) {
      setPlayed(state.played)
    }
  }

  const playPauseHandler = (): void => {
    setPlaying(!playing)
  }

  const seekHandler = (_e: Event, value: number): void => {
    const newPlayed = value / 100
    setPlayed(newPlayed)
    videoPlayerRef?.current?.seekTo(newPlayed)
  }

  const seekMouseUpHandler = (_e: SyntheticEvent<Element, Event>, value: number): void => {
    setSeeking(false)
    videoPlayerRef?.current?.seekTo(value / 100)
  }

  const volumeChangeHandler = (_e: Event, value: number): void => {
    setVolume(value / 100)
    setMuted(value === 0)
  }

  const volumeSeekUpHandler = (e: SyntheticEvent<Element, Event>, value: number): void => {
    volumeChangeHandler(e.nativeEvent, value)
  }

  const muteHandler = (): void => {
    // TODO: verify volume is not 0
    setMuted(!muted)
  }

  const onSeekMouseDownHandler = (): void => {
    setSeeking(true)
  }

  const mouseMoveHandler = (): void => {
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
            <LoadFileButton setFileUrl={setUrl} fileName={props.videoName} />
          </div>
          )}
    </>
  )
}
