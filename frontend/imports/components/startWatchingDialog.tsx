/** @jsxImportSource @emotion/react */
import React, { ChangeEvent, useState } from 'react'
import {
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Button,
  TextField,
  Typography
} from '@mui/material'
import config from '../../config.json'
import { useRouter } from 'next/router'
import { css } from '@emotion/react'

const onEnter = (func: () => void | Promise<void>) => (e: React.KeyboardEvent<HTMLDivElement>) => {
  if (e.key === 'Enter') return func()
}

const StartWatchingDialog = (props: { shown: boolean, handleClose: () => void }) => {
  const [title, setTitle] = useState('')
  const [fileName, setFileName] = useState('')
  const [fileUrl, setFileUrl] = useState<URL | null>(null)
  const [error, setError] = useState('')
  const [inProgress, setInProgress] = useState(false)

  const router = useRouter()

  const handleClose = () => {
    props.handleClose()
  }

  const handleCreateRoom = async () => {
    setInProgress(true)
    try {
      const req = await fetch(config.serverUrl + '/api/room', {
        method: 'POST',
        body: JSON.stringify({ title })
      })
      const res: { error?: string, id: string } = await req.json()
      if (res.error) setError(res.error)
      else {
        props.handleClose()
        router.push(`/room/${res.id}`).catch(console.error)
      }
    } catch (e) { setError('An unknown network error occurred.') }
    setInProgress(false)
  }

  const handleFileSelect = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files?.length !== 1) {
      return
    }
    const file = e.target.files[0]
    setFileName(file.name)

    const url = URL.createObjectURL(file)
    setFileUrl(new URL(url))
  }

  const createButtonDisabled = !title && !fileName
  return (
    <Dialog open={props.shown} onClose={handleClose}>
      <DialogTitle>Create a Room</DialogTitle>

      <DialogContent css={{ paddingBottom: 0 }}>
        <TextField
          value={title} onChange={e => setTitle(e.target.value)}
          onKeyDown={onEnter(handleCreateRoom)}
          margin='dense' label='Title' type='text' fullWidth
        />

        <div css={css`
          display: flex;
          flex-direction: row;
          align-items: center;
          margin-top: 16px;
          min-width: 400px;
        `}
        >
          <Button
            component='label'
            variant='outlined'
            css={css`
              margin-right: 8px;
            `}
          >
            Select Video
            <input type='file' hidden onChange={handleFileSelect} />
          </Button>

          <Typography>{fileName}</Typography>
        </div>

        <Typography color='error' css={{ marginTop: 8 }} gutterBottom>{error}</Typography>
      </DialogContent>

      <DialogActions>
        <Button onClick={handleCreateRoom} disabled={createButtonDisabled || inProgress}>
          Create
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default StartWatchingDialog
