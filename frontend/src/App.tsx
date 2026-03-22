import { RouterProvider } from 'react-router-dom';
import { router } from './router';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN'; // 使用中文包

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <RouterProvider router={router} />
    </ConfigProvider>
  );
}

export default App;