import joinWebm from '$lib/assets/join.webm'
import leaveWebm from '$lib/assets/leave.webm'
import messageWebm from '$lib/assets/message.webm'

export const join = typeof Audio !== 'undefined' ? new Audio(joinWebm) : null
export const leave = typeof Audio !== 'undefined' ? new Audio(leaveWebm) : null
export const message = typeof Audio !== 'undefined' ? new Audio(messageWebm) : null
