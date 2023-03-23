import React from 'react'
import BaseReactPlayer, { BaseReactPlayerProps } from 'react-player/base'
import ReactPlayer, { ReactPlayerProps } from 'react-player/lazy'

const ReactPlayerWrapper = (props: ReactPlayerProps & { playerRef: React.Ref<BaseReactPlayer<BaseReactPlayerProps>> }) => {
  return (
    <ReactPlayer
      ref={props.playerRef}
      {...props}
    />
  )
}

export default ReactPlayerWrapper
