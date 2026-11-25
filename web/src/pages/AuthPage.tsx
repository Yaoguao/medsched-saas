import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { Activity, ArrowRight, LockKeyhole, UserPlus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { authApi } from '../shared/api/auth'
import { setAccessToken } from '../shared/lib/auth'
import { Button } from '../shared/ui/Button'
import { TextField } from '../shared/ui/TextField'

type AuthMode = 'login' | 'register'

const loginSchema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(8, 'Минимум 8 символов'),
})

const registerSchema = loginSchema.extend({
  phone: z.string().optional(),
})

type LoginForm = z.infer<typeof loginSchema>
type RegisterForm = z.infer<typeof registerSchema>

type AuthPageProps = {
  initialMode: AuthMode
}

export function AuthPage({ initialMode }: AuthPageProps) {
  const [mode, setMode] = useState<AuthMode>(initialMode)

  return (
    <main className="auth-shell">
      <section className="auth-brand">
        <div className="brand-mark">
          <Activity size={26} />
        </div>
        <div>
          <p className="eyebrow">MedSched SaaS</p>
          <h1>Вход в систему</h1>
          <p className="lead">
            Рабочее пространство клиники для врачей, регистраторов, пациентов
            и управленческой аналитики.
          </p>
        </div>
      </section>

      <section className="auth-panel">
        <div className="auth-tabs">
          <Link
            className={mode === 'login' ? 'active' : ''}
            to="/login"
            onClick={() => setMode('login')}
          >
            Вход
          </Link>
          <Link
            className={mode === 'register' ? 'active' : ''}
            to="/register"
            onClick={() => setMode('register')}
          >
            Регистрация
          </Link>
        </div>

        {mode === 'login' ? <LoginFormView /> : <RegisterFormView />}
      </section>
    </main>
  )
}

function LoginFormView() {
  const navigate = useNavigate()
  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: 'patient@example.com',
      password: 'password123',
    },
  })
  const mutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      setAccessToken(data.access_token)
      navigate('/home')
    },
  })

  return (
    <form className="auth-form" onSubmit={form.handleSubmit((data) => mutation.mutate(data))}>
      <div>
        <p className="eyebrow">Авторизация</p>
        <h2>Войти в систему</h2>
      </div>
      <TextField
        label="Email"
        type="email"
        autoComplete="email"
        error={form.formState.errors.email?.message}
        {...form.register('email')}
      />
      <TextField
        label="Пароль"
        type="password"
        autoComplete="current-password"
        error={form.formState.errors.password?.message}
        {...form.register('password')}
      />
      {mutation.error ? <p className="form-error">{mutation.error.message}</p> : null}
      <Button
        type="submit"
        icon={<LockKeyhole size={18} />}
        disabled={mutation.isPending}
      >
        {mutation.isPending ? 'Входим...' : 'Войти'}
      </Button>
    </form>
  )
}

function RegisterFormView() {
  const navigate = useNavigate()
  const form = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: 'patient@example.com',
      phone: '+79990000999',
      password: 'password123',
    },
  })
  const mutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: () => {
      navigate('/login')
    },
  })

  return (
    <form className="auth-form" onSubmit={form.handleSubmit((data) => mutation.mutate(data))}>
      <div>
        <p className="eyebrow">Новый пользователь</p>
      </div>
      <TextField
        label="Email"
        type="email"
        autoComplete="email"
        error={form.formState.errors.email?.message}
        {...form.register('email')}
      />
      <TextField label="Телефон" autoComplete="tel" {...form.register('phone')} />
      <TextField
        label="Пароль"
        type="password"
        autoComplete="new-password"
        error={form.formState.errors.password?.message}
        {...form.register('password')}
      />
      {mutation.error ? <p className="form-error">{mutation.error.message}</p> : null}
      <Button
        type="submit"
        icon={<UserPlus size={18} />}
        disabled={mutation.isPending}
      >
        {mutation.isPending ? 'Создаем...' : 'Зарегистрировать'}
      </Button>
      {mutation.isSuccess ? (
        <p className="form-success">
          Аккаунт создан. Профиль пациента привяжет регистратор после создания карты.
          <ArrowRight size={16} />
        </p>
      ) : null}
    </form>
  )
}
