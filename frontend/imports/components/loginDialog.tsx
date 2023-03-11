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

const onEnter = <T,>(func: () => T) => (e: React.KeyboardEvent<HTMLDivElement>) => {
  if (e.key === 'Enter') return func()
}

const LoginDialog = (props: { shown: boolean, handleClose: () => void }) => {
  const setLoginStatus = useSetRecoilState(loginStatusAtom)
  const passwordRef = useRef<HTMLInputElement>()
  const confirmRef = useRef<HTMLInputElement>()
  const emailRef = useRef<HTMLInputElement>()
  const [passwordMatches, setPasswordMatches] = useState(false)
  const [registerMode, setRegisterMode] = useState(false)
  const [inProgress, setInProgress] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')

  const emailPresent = !registerMode || (registerMode && email)
  const errorColor = error.startsWith('Your account has been created!') ? 'primary' : 'error'

  const clearDialog = () => {
    setPasswordMatches(false)
    setInProgress(false)
    setUsername('')
    setPassword('')
    setEmail('')
    setError('')
  }

  const handleClose = () => {
    setRegisterMode(false)
    clearDialog()
    props.handleClose()
  }

  const handleRegister = () => {
    setRegisterMode(state => !state)
    clearDialog()
  }

  const submit = async () => {
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
        else if (res.success) {
          setError('Your account has been created! Wait for an admin to verify you.')
        } else {
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

  const handleSubmit = () => { submit().catch(console.error) }

  const usernameFieldLabel = (!registerMode ? 'Email Address/' : '') + 'Username'
  const loginButtonDisabled = !username || !password || !emailPresent
  return (
    <Dialog open={props.shown} onClose={handleClose}>
      <DialogTitle>{registerMode ? 'Register' : 'Login'}</DialogTitle>

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
          onKeyDown={onEnter(() => registerMode ? confirmRef.current?.focus() : handleSubmit())}
          margin='dense' label='Password' type='password' fullWidth inputRef={passwordRef}
        />
        {registerMode && (
          <TextField
            onChange={e => setPasswordMatches(e.target.value === password)}
            onKeyDown={onEnter(() => handleSubmit())}
            margin='dense' label='Confirm Password' type='password' fullWidth inputRef={confirmRef}
          />
        )}
        {!registerMode && (
          <Typography css={{ marginTop: 8 }} gutterBottom color='primary'>
            Forgot your password? Contact the site admins.
          </Typography>
        )}
        <Typography color={errorColor} css={{ marginTop: 8 }} gutterBottom>{error}</Typography>
      </DialogContent>

      <DialogActions>
        <Button onClick={handleRegister} color='secondary' disabled={inProgress}>
          {registerMode ? 'Login' : 'Register'}
        </Button>
        <Button onClick={handleSubmit} disabled={loginButtonDisabled || inProgress}>
          {registerMode ? 'Register' : 'Login'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default LoginDialog
