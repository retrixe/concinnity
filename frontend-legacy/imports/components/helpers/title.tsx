import React from 'react'
import Head from 'next/head'

const Title = (props: {
  title: string
  description: string
  url: string
  noIndex?: boolean
}): React.JSX.Element => (
  <Head>
    <title>{props.title}</title>
    <meta property='og:title' content={props.title} />
    <meta property='og:url' content={props.url} />
    <meta property='og:description' content={props.description} />
    <meta name='Description' content={props.description} />
    {props.noIndex && <meta name='robots' content='noindex,nofollow' />}
  </Head>
)

export default Title
