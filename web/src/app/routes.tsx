import { lazy, Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { ContentState } from '../shared/ui/ContentState'

const AuthPage = lazy(() => import('../pages/AuthPage').then((module) => ({ default: module.AuthPage })))
const AdminDashboardPage = lazy(() =>
  import('../pages/AdminDashboardPage').then((module) => ({ default: module.AdminDashboardPage })),
)
const AppointmentDetailsPage = lazy(() =>
  import('../pages/AppointmentDetailsPage').then((module) => ({ default: module.AppointmentDetailsPage })),
)
const BookingPage = lazy(() => import('../pages/BookingPage').then((module) => ({ default: module.BookingPage })))
const HomePage = lazy(() => import('../pages/HomePage').then((module) => ({ default: module.HomePage })))
const ProfilePage = lazy(() => import('../pages/ProfilePage').then((module) => ({ default: module.ProfilePage })))

export function AppRoutes() {
  return (
    <Suspense fallback={<main className="route-loading"><ContentState variant="loading" title="Загружаю раздел" /></main>}>
      <Routes>
        <Route path="/" element={<Navigate to="/home" replace />} />
        <Route path="/home" element={<HomePage />} />
        <Route path="/booking" element={<BookingPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/login" element={<AuthPage initialMode="login" />} />
        <Route path="/register" element={<AuthPage initialMode="register" />} />
        <Route path="/workspace" element={<AdminDashboardPage />} />
        <Route path="/workspace/appointments/:appointmentId" element={<AppointmentDetailsPage />} />
        <Route path="/admin" element={<Navigate to="/workspace" replace />} />
        <Route path="*" element={<Navigate to="/home" replace />} />
      </Routes>
    </Suspense>
  )
}
