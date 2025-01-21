// https://github.com/silviapfeiffer/silviapfeiffer.github.io/blob/master/index.html

export function srt2webvtt(data: string) {
  // remove dos newlines
  let srt = data.replace(/\r+/g, '')
  // trim white space start and end
  srt = srt.replace(/^\s+|\s+$/g, '')

  // get cues
  const cuelist = srt.split('\n\n')
  let result = ''

  if (cuelist.length > 0) {
    result += 'WEBVTT\n\n'
    for (const cue of cuelist) {
      result += convertSrtCue(cue)
    }
  }

  return result
}

function convertSrtCue(caption: string) {
  // remove all html tags for security reasons
  //srt = srt.replace(/<[a-zA-Z\/][^>]*>/g, '');

  let cue = ''
  const s = caption.split(/\n/)

  // concatenate muilt-line string separated in array into one
  while (s.length > 3) {
    for (let i = 3; i < s.length; i++) {
      s[2] += '\n' + s[i]
    }
    s.splice(3, s.length - 3)
  }

  let line = 0

  // detect identifier
  if (!/\d+:\d+:\d+/.exec(s[0]) && /\d+:\d+:\d+/.exec(s[1])) {
    cue += /\w+/.exec(s[0]) + '\n' // eslint-disable-line
    line += 1
  }

  // get time strings
  if (/\d+:\d+:\d+/.exec(s[line])) {
    // convert time string
    const m = /(\d+):(\d+):(\d+)(?:,(\d+))?\s*--?>\s*(\d+):(\d+):(\d+)(?:,(\d+))?/.exec(s[1])
    if (m) {
      cue +=
        m[1] +
        ':' +
        m[2] +
        ':' +
        m[3] +
        '.' +
        m[4] +
        ' --> ' +
        m[5] +
        ':' +
        m[6] +
        ':' +
        m[7] +
        '.' +
        m[8] +
        '\n'
      line += 1
    } else {
      // Unrecognized timestring
      return ''
    }
  } else {
    // file format error or comment lines
    return ''
  }

  // get cue text
  if (s[line]) {
    cue += s[line] + '\n\n'
  }

  return cue
}
