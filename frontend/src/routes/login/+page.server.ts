import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    // TODO: Metadata
    title: 'Sign In - concinnity',
    image:
      'https://media.discordapp.net/attachments/588340346841464835/1321795849571008572/image.png?ex=67708410&is=676f3290&hm=7d04e84e556d48740664a0b5368009b0c21e73a4037b896a7920bbbb6cc7a0bf&=&format=webp&quality=lossless&width=1536&height=844',
    description:
      'Watch videos with your friends together using concinnity, a FOSS, lightweight and easy to use website built by a developer looking for something better.',
  }
}
