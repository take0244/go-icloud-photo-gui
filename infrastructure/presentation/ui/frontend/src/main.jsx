import React from 'react'
import {createRoot} from 'react-dom/client'
import App from './App'
import { transitions, positions, Provider as AlertProvider } from 'react-alert'
import AlertTemplate from 'react-alert-template-basic'

const container = document.getElementById('root')

const root = createRoot(container)
const options = {
  position: positions.TOP_RIGHT,
  timeout: 5000,
  offset: '30px',
  transition: transitions.SCALE
}


root.render(
  <React.StrictMode>
    <AlertProvider template={AlertTemplate} {...options}>
      <App />
    </AlertProvider>
  </React.StrictMode>
)
