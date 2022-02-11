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

const onEnter = (func: () => void | Promise<void>) => (e: React.KeyboardEvent<HTMLDivElement>) => {
  if (e.key === 'Enter') return func()
}

const LoginDialog = (props: { shown: boolean, handleClose: () => void }) => {
  const setLoginStatus = useSetRecoilState(loginStatusAtom)
  const passwordRef = useRef<HTMLInputElement>()
  const confirmRef = useRef<HTMLInputElement>()
  const emailRef = useRef<HTMLInputElement>()
  const [inProgress, setInProgress] = useState(false)
  const [passwordMatches, setPasswordMatches] = useState(false)
  const [registerMode, setRegisterMode] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')

  const emailPresent = !registerMode || (registerMode && email)

  const handleClose = () => {
    setPasswordMatches(false)
    setRegisterMode(false)
    setInProgress(false)
    setUsername('')
    setPassword('')
    setEmail('')
    setError('')
    props.handleClose()
  }
  const handleRegister = () => setRegisterMode(state => !state)
  const handleLoginDialog = async () => {
    if (registerMode && !passwordMatches) setError('Your entered passwords don\'t match!')
    else if (username && password && emailPresent) {
      setInProgress(true)
      const endpoint = registerMode ? '/api/register' : '/api/login'
      try {
        const req = await fetch(config.serverUrl + endpoint, {
          method: 'POST',
          body: JSON.stringify({ username, password, email })
        })
        const res = await req.json()
        if (res.error) setError(res.error)
        else {
          localStorage.setItem('token', res.token)
          setLoginStatus(res.username)
          handleClose()
        }
        setInProgress(false)
      } catch (e) { setError('An unknown network error occurred.') }
    } else if (!username) setError('Enter your username' + (registerMode ? '' : 'or e-mail') + '!')
    else if (!password) setError('Enter your password!')
    else if (!emailPresent) setError('Enter your e-mail!')
  }

  const usernameFieldLabel = (!registerMode ? 'Email Address/' : '') + 'Username'
  const loginButtonDisabled = !username || !password || !emailPresent
  return (
    <Dialog open={props.shown} onClose={handleClose}>
      <DialogTitle>Login</DialogTitle>

      <DialogContent css={{ paddingBottom: 0 }}>
        <TextField
          value={username} onChange={e => setUsername(e.target.value)}
          onKeyDown={onEnter(() => (registerMode ? emailRef : passwordRef).current?.focus())}
          margin='dense' label={usernameFieldLabel} type='email' fullWidth autoFocus
        />
        {registerMode && (
          <TextField
            value={email} onChange={e => setEmail(e.target.value)}
            onKeyDown={onEnter(() => passwordRef.current?.focus())}
            margin='dense' label='Email Address' type='email' fullWidth inputRef={emailRef}
          />
        )}
        <TextField
          value={password} onChange={e => setPassword(e.target.value)}
          onKeyDown={onEnter(() => registerMode ? confirmRef.current?.focus() : handleLoginDialog())}
          margin='dense' label='Password' type='password' fullWidth inputRef={passwordRef}
        />
        {registerMode && (
          <TextField
            onChange={e => setPasswordMatches(e.target.value === password)}
            onKeyDown={onEnter(async () => await handleLoginDialog())}
            margin='dense' label='Confirm Password' type='password' fullWidth inputRef={confirmRef}
          />
        )}
        <Typography color='error' css={{ marginTop: 8 }} gutterBottom>{error}</Typography>
      </DialogContent>

      <DialogActions>
        <Button onClick={handleRegister} color='secondary' disabled={inProgress}>
          {registerMode ? 'Login' : 'Register'}
        </Button>
        <Button onClick={handleLoginDialog} disabled={loginButtonDisabled || inProgress}>
          {registerMode ? 'Register' : 'Login'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default LoginDialog
