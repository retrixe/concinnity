import kyLibrary from 'ky'
import { PUBLIC_BACKEND_URL } from '$env/static/public'

const ky = kyLibrary.create({
  prefixUrl: PUBLIC_BACKEND_URL,
  hooks: {
    beforeRequest: [
      req => req.headers.set('Authorization', localStorage.getItem('concinnity:token') ?? ''),
    ],
    beforeError: [
      async error => {
        try {
          // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
          const data = await error.response?.json<{ error?: string }>()
          if (data.error) {
            error.name = 'ConcinnityError'
            error.message = data.error
          }
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
        } catch (e: unknown) {
          /* Do nothing */
        }

        return error
      },
    ],
  },
})

export default ky
