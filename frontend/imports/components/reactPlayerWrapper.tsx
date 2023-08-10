import React from 'react'
import { type BaseReactPlayerProps } from 'react-player/base'
import type BaseReactPlayer from 'react-player/base'
import ReactPlayer, { type ReactPlayerProps } from 'react-player/lazy'

const ReactPlayerWrapper = (
  props: ReactPlayerProps & { playerRef: React.Ref<BaseReactPlayer<BaseReactPlayerProps>> }
): JSX.Element => (
  <ReactPlayer
    ref={props.playerRef}
    {...props}
  />
)

export default ReactPlayerWrapper
