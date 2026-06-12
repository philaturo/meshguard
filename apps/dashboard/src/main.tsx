// File: apps/dashboard/src/main.tsx
// Purpose: React application entry point
// Connects to: App.tsx (root component), index.html (DOM mount point)

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
