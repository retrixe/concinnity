import kyLibrary, { isHTTPError } from 'ky'
import { PUBLIC_BACKEND_URL } from '$env/static/public'

const ky = kyLibrary.create({
  prefix: PUBLIC_BACKEND_URL,
  hooks: {
    beforeRequest: [
      ({ request }) =>
        request.headers.set('Authorization', localStorage.getItem('concinnity:token') ?? ''),
    ],
    beforeError: [
      ({ error }) => {
        try {
          if (isHTTPError(error)) {
            const data: unknown =
              typeof error.data === 'string' ? JSON.parse(error.data) : error.data
            if (
              typeof data === 'object' &&
              data !== null &&
              'error' in data &&
              typeof data.error === 'string'
            ) {
              // @ts-expect-error -- We're transforming the existing error, it's fine
              error.name = 'ConcinnityError'
              error.message = data.error
            }
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
