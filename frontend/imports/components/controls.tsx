import { IconButton, Slider, Typography } from '@mui/material'
import { Pause, PlayArrow, VolumeOff, VolumeUp } from '@mui/icons-material'
import React, { SyntheticEvent } from 'react'

interface Props {
  videoName: string
  onPlayPause: () => void
  playing: boolean
  played: number
  onSeek: (e: Event, newValue: number) => void
  onMouseSeekUp: (e: SyntheticEvent<Element, Event>, newValue: number) => void
  onVolumeChangeHandler: (e: Event, newValue: number) => void
  onVolumeSeekUp: (e: SyntheticEvent<Element, Event>, newValue: number) => void
  volume: number
  mute: boolean
  onMute: () => void
  duration: string
  currentTime: string
  onMouseSeekDown: () => void
  controlRef: React.RefObject<HTMLDivElement>
  // TODO: subs
}

export const Controls = (props: Props) => {
  return (
    <div ref={props.controlRef}>
      <div>
        <Typography variant='h6'>{props.videoName}</Typography>
      </div>
      <div>
        <IconButton onClick={props.onPlayPause}>
          {props.playing ? <Pause /> : <PlayArrow />}
        </IconButton>
      </div>
      <div>
        <div>
          <Slider
            min={0}
            max={100}
            value={props.played * 100}
            onChange={(e, number) => props.onSeek(e, number as number)}
            onChangeCommitted={(e, number) =>
              props.onMouseSeekUp(e as SyntheticEvent<Element, Event>, number as number)}
            onMouseDown={() => props.onMouseSeekDown()}
            component='div'
          />
        </div>
        <div>
          <IconButton onClick={props.onPlayPause}>
            {props.playing ? <Pause /> : <PlayArrow />}
          </IconButton>
          <IconButton onClick={props.onMute}>
            {props.mute ? <VolumeOff /> : <VolumeUp />}
          </IconButton>
        </div>
        <Slider
          value={props.volume * 100}
          onChange={(e, number) => props.onVolumeChangeHandler(e, number as number)}
          onChangeCommitted={(e, number) =>
            props.onVolumeSeekUp(e as SyntheticEvent<Element, Event>, number as number)}
          component='div'
        />
        <span>{props.currentTime}/{props.duration}</span>
      </div>
    </div>
  )
}
