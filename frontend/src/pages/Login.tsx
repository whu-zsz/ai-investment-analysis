import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { useState } from 'react';

interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

const Login = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const { register, handleSubmit, formState: { errors } } = useForm<LoginForm>();

  const onSubmit = async (data: LoginForm) => {
    setLoading(true);
    setError('');

    // 模拟登录（后面改成调用后端 /api/v1/login 接口）
    setTimeout(() => {
      if (data.username === '盛哲' && data.password === '123456') {
        localStorage.setItem('token', 'fake-jwt-token-2026');
        localStorage.setItem('username', data.username);
        navigate('/'); // 跳转首页
      } else {
        setError('用户名或密码错误（测试账号：盛哲 / 123456）');
      }
      setLoading(false);
    }, 800);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
      <div className="bg-white rounded-3xl shadow-2xl w-full max-w-md p-10">
        <div className="text-center mb-8">
          <div className="mx-auto w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center text-white text-4xl mb-4">
            AI
          </div>
          <h1 className="text-3xl font-bold">AI投顾助手</h1>
          <p className="text-gray-500 mt-2">武汉大学计算机学院课程设计</p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">用户名 / 学号</label>
            <input
              {...register('username', { required: '用户名不能为空' })}
              className="w-full px-5 py-3 border border-gray-300 rounded-2xl focus:outline-none focus:border-blue-500"
              placeholder="例如：盛哲 或 2023302111118"
            />
            {errors.username && <p className="text-red-500 text-sm mt-1">{errors.username.message}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">密码</label>
            <input
              type="password"
              {...register('password', { required: '密码不能为空' })}
              className="w-full px-5 py-3 border border-gray-300 rounded-2xl focus:outline-none focus:border-blue-500"
              placeholder="输入密码"
            />
            {errors.password && <p className="text-red-500 text-sm mt-1">{errors.password.message}</p>}
          </div>

          <div className="flex items-center justify-between">
            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" {...register('remember')} className="w-4 h-4" />
              记住我
            </label>
            <a href="#" className="text-blue-600 text-sm hover:underline">忘记密码？</a>
          </div>

          {error && <p className="text-red-500 text-center">{error}</p>}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-4 bg-blue-600 text-white font-semibold rounded-2xl hover:bg-blue-700 transition-all disabled:opacity-70"
          >
            {loading ? '登录中...' : '立即登录'}
          </button>
        </form>

        <p className="text-center text-xs text-gray-400 mt-8">
          测试账号：盛哲 / 123456<br />
          （实际项目请对接后端 /login 接口）
        </p>
      </div>
    </div>
  )
}

export default Login