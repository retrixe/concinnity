import React from 'react'
import Title from '../imports/components/title'
import { AppDiv, TopBar } from '../imports/components/layout'

const IndexPage = () => {
  return (
    <>
      <Title title='Home - Concinnity' url='/' description='' />
      <TopBar />
      <AppDiv>
        <p>Hello, world!</p>
      </AppDiv>
    </>
  )
}

export default IndexPage
