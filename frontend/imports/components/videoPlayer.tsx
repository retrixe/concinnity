
import { css } from '@emotion/react'
import React, { useState } from 'react'
import dynamic from 'next/dynamic'
import { Button } from '@mui/material'
// Fix for Hydration error
const ReactPlayer = dynamic(async () => await import('react-player'), { ssr: false })

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

export const VideoPlayer = (props: { url?: string }) => {
  const [url, setUrl] = useState(props.url)

  return (
    <>
      {url
        ? (
          <ReactPlayer
            css={css`
              background-color: #000;
              margin: 10px 0;
            `}
            url={url}
            width={800}
            height={450}
            controls
          />)
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
