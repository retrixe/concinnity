/** @jsxImportSource @emotion/react */
import React, { useRef, useState } from 'react'
import { useSetRecoilState } from 'recoil'
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
import { loginStatusAtom } from '../recoil-atoms'

const LoginDialog = (props: { shown: boolean, handleClose: () => void }) => {
  const setLoginStatus = useSetRecoilState(loginStatusAtom)
  const passwordRef = useRef<HTMLInputElement>()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const handleClose = () => {
    setUsername('')
    setPassword('')
    setError('')
    props.handleClose()
  }
  const handleLoginDialog = async () => {
    if (username && password) {
      try {
        const req = await fetch(config.serverUrl + '/api/login', {
          method: 'POST',
          body: JSON.stringify({ username, password })
        })
        const res = await req.json()
        if (res.error) setError(res.error)
        else {
          localStorage.setItem('token', res.token)
          setLoginStatus(res.username)
          handleClose()
        }
      } catch (e) { setError('An unknown network error occurred.') }
    } else if (!username) setError('Enter a username or e-mail.')
    else if (!password) setError('Enter a password.')
  }
  return (
    <Dialog open={props.shown} onClose={handleClose}>
      <DialogTitle>Login</DialogTitle>
      <DialogContent css={{ paddingBottom: 0 }}>
        <TextField
          value={username} onChange={e => setUsername(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && passwordRef.current?.focus()}
          margin='dense' label='Email Address/Username' type='email' fullWidth autoFocus
        />
        <TextField
          value={password} onChange={e => setPassword(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleLoginDialog()}
          margin='dense' label='Password' type='password' fullWidth inputRef={passwordRef}
        />
        <Typography color='error' css={{ marginTop: 8 }} gutterBottom>{error}</Typography>
      </DialogContent>
      <DialogActions>
        <Button disabled>Register (N/A)</Button>
        <Button onClick={handleLoginDialog}>Login</Button>
      </DialogActions>
    </Dialog>
  )
}

export default LoginDialog
