// src/main.tsx 
import React from 'react'
import ReactDOM from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import { router } from './router' // 确保导入了组员修改的那个 router

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    {/* 必须使用 RouterProvider，否则无法识别 /app 路径 */}
    <RouterProvider router={router} />
  </React.StrictMode>
)