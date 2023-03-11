import { useRouter } from 'next/router'
import React from 'react'
import config from '../../config.json'

const RoomPage = () => {
  const id = useRouter().query.id

  React.useEffect(() => {
    if (typeof id !== 'string' || !id) return
    fetch(config.serverUrl + `/api/room/${id}`, {
      method: 'GET',
      headers: { Authentication: localStorage.getItem('token') ?? '' }
    }).then(async res => await res.json()).then(json => {
      if (json.error) {
        console.error(json.error)
      } else {
        console.log(json)
      }
    }).catch(console.error)
  }, [id])

  return (
    <p>Hello, world!</p>
  )
}

export default RoomPage
