import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './i18n' // i18n 초기화 (App 전에 실행)
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
