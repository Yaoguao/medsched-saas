import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Activity, ArrowLeft, HeartPulse, LogOut, Stethoscope, UserRound } from 'lucide-react'
import type { FormEvent } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { authApi } from '../shared/api/auth'
import { coreApi } from '../shared/api/core'
import type { MedicalProfile, Patient, UpdatePatientPayload } from '../shared/api/types'
import { clearAccessToken, getAccessToken } from '../shared/lib/auth'
import { userRoleLabel, userStatusLabel } from '../shared/lib/labels'
import { getPatientMedicalProfile } from '../shared/lib/medicalProfile'
import { Button } from '../shared/ui/Button'

export function ProfilePage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const token = getAccessToken()
  const meQuery = useQuery({
    queryKey: ['me'],
    queryFn: authApi.me,
    enabled: Boolean(token),
  })
  const patientQuery = useQuery({
    queryKey: ['profile-patient', meQuery.data?.patient_id],
    queryFn: () => coreApi.patient(meQuery.data?.patient_id ?? ''),
    enabled: Boolean(token && meQuery.data?.patient_id),
  })
  const doctorQuery = useQuery({
    queryKey: ['profile-doctor', meQuery.data?.doctor_id],
    queryFn: () => coreApi.doctor(meQuery.data?.doctor_id ?? ''),
    enabled: Boolean(token && meQuery.data?.doctor_id),
  })
  const patientUpdateMutation = useMutation({
    mutationFn: (payload: UpdatePatientPayload) => coreApi.updatePatient(meQuery.data?.patient_id ?? '', payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profile-patient', meQuery.data?.patient_id] })
      queryClient.invalidateQueries({ queryKey: ['home-patient', meQuery.data?.patient_id] })
    },
  })

  if (!token) {
    return <Navigate to="/login" replace />
  }

  const user = meQuery.data
  const patient = patientQuery.data
  const medicalProfile = getPatientMedicalProfile(patient)
  const doctor = doctorQuery.data

  return (
    <main className="home-shell">
      <header className="home-header">
        <div className="home-brand">
          <div className="brand-mark small">
            <Activity size={20} />
          </div>
          <div>
            <strong>MedSched</strong>
            <span>Профиль</span>
          </div>
        </div>
        <div className="doctor-chip">
          <UserRound size={18} />
          <span>{user?.email ?? 'Пользователь'}</span>
          <Button variant="secondary" icon={<ArrowLeft size={16} />} onClick={() => navigate('/home')}>
            Главная
          </Button>
          <Button
            variant="ghost"
            icon={<LogOut size={16} />}
            onClick={() => {
              clearAccessToken()
              navigate('/login')
            }}
          >
            Выйти
          </Button>
        </div>
      </header>

      <section className="home-hero">
        <div>
          <p className="eyebrow">{user?.role ? userRoleLabel(user.role) : 'Профиль'}</p>
          <h1>{profileHeading(patient?.full_name, doctor?.full_name, user?.email)}</h1>
          <p className="lead">
            Здесь отображаются учетная запись и связанные профили пациента или врача.
          </p>
        </div>
        <div className="home-status">
          <span>Статус</span>
          <strong>{userStatusLabel(user?.status ?? 'ACTIVE')}</strong>
          <small>{user?.role ? userRoleLabel(user.role) : '-'}</small>
        </div>
      </section>

      <section className="profile-page-grid">
        <article className="panel">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Аккаунт</p>
              <h2>Учетная запись</h2>
            </div>
          </div>
          <div className="profile-summary">
            <ProfileRow label="Email" value={user?.email} />
            <ProfileRow label="Телефон" value={user?.phone} />
            <ProfileRow label="Роль" value={user?.role ? userRoleLabel(user.role) : undefined} />
            <ProfileRow label="ID аккаунта" value={user?.id} />
          </div>
        </article>

        {patient ? (
          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Пациент</p>
                <h2>Профиль пациента</h2>
              </div>
              <HeartPulse size={20} />
            </div>
            <PatientProfileForm
              patient={patient}
              isPending={patientUpdateMutation.isPending}
              onSubmit={(payload) => patientUpdateMutation.mutate(payload)}
            />
            {patientUpdateMutation.isError ? (
              <p className="form-error">{patientUpdateMutation.error.message}</p>
            ) : null}
            {patientUpdateMutation.isSuccess ? (
              <p className="form-success">Профиль обновлен.</p>
            ) : null}
          </article>
        ) : null}

        {medicalProfile ? (
          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Медицинский профиль</p>
                <h2>Медицинские признаки</h2>
              </div>
            </div>
            <MedicalProfileView profile={medicalProfile} />
            <p className="section-note">
              Эти признаки редактирует медперсонал или регистратор по данным карты пациента.
            </p>
          </article>
        ) : null}

        {doctor ? (
          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Doctor</p>
                <h2>Профиль врача</h2>
              </div>
              <Stethoscope size={20} />
            </div>
            <div className="profile-summary">
              <ProfileRow label="ФИО" value={doctor.full_name} />
              <ProfileRow label="Специальность" value={doctor.specialty} />
              <ProfileRow label="Кабинет" value={doctor.cabinet} />
              <ProfileRow label="Стаж" value={String(doctor.experience_years)} />
              <ProfileRow label="Doctor ID" value={doctor.id} />
            </div>
          </article>
        ) : null}

        {!patient && !doctor && !meQuery.isLoading ? (
          <article className="panel">
            <p className="empty-state">
              К учетной записи пока не привязан профиль пациента или врача.
              <Link to="/home"> Вернуться на главную.</Link>
            </p>
          </article>
        ) : null}
      </section>
    </main>
  )
}

function ProfileRow({ label, value }: { label: string; value?: string }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value || '-'}</strong>
    </div>
  )
}

function PatientProfileForm({
  patient,
  isPending,
  onSubmit,
}: {
  patient: Patient
  isPending: boolean
  onSubmit: (payload: UpdatePatientPayload) => void
}) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onSubmit({
      full_name: String(form.get('full_name') ?? ''),
      gender: String(form.get('gender') ?? ''),
      birth_date: `${String(form.get('birth_date') ?? '')}T00:00:00Z`,
      age: Number(form.get('age') ?? 0),
      phone: String(form.get('phone') ?? ''),
      email: String(form.get('email') ?? ''),
      neighbourhood: String(form.get('neighbourhood') ?? ''),
      scholarship: form.get('scholarship') === 'on',
    })
  }

  return (
    <form className="admin-form" onSubmit={submit}>
      <input name="full_name" defaultValue={patient.full_name} placeholder="ФИО" required />
      <select name="gender" defaultValue={patient.gender}>
        <option value="F">Женский</option>
        <option value="M">Мужской</option>
      </select>
      <input name="birth_date" type="date" defaultValue={patient.birth_date.slice(0, 10)} />
      <input name="age" type="number" min="0" defaultValue={patient.age} placeholder="Возраст" />
      <input name="phone" defaultValue={patient.phone} placeholder="Телефон" />
      <input name="email" type="email" defaultValue={patient.email} placeholder="Email" />
      <input name="neighbourhood" defaultValue={patient.neighbourhood} placeholder="Район" />
      <label className="checkbox-line">
        <input name="scholarship" type="checkbox" defaultChecked={patient.scholarship} />
        Льгота / scholarship
      </label>
      <Button type="submit" disabled={isPending}>Сохранить профиль</Button>
      <div className="profile-summary compact">
        <ProfileRow label="ID пациента" value={patient.id} />
      </div>
    </form>
  )
}

function MedicalProfileView({ profile }: { profile: MedicalProfile }) {
  return (
    <div className="profile-summary">
      <ProfileRow label="Хронические заболевания" value={profile.chronic_conditions} />
      <ProfileRow label="Факторы риска" value={profile.risk_factors} />
      <ProfileRow label="Аллергии" value={profile.allergies} />
      <ProfileRow label="Заметки" value={profile.notes} />
      <ProfileRow label="Гипертония" value={profile.has_hypertension ? 'Да' : 'Нет'} />
      <ProfileRow label="Диабет" value={profile.has_diabetes ? 'Да' : 'Нет'} />
      <ProfileRow label="Алкоголизм" value={profile.has_alcoholism ? 'Да' : 'Нет'} />
      <ProfileRow label="Инвалидность" value={profile.has_handicap ? 'Да' : 'Нет'} />
    </div>
  )
}

function profileHeading(patientName?: string, doctorName?: string, email?: string) {
  return doctorName ?? patientName ?? email ?? 'Профиль пользователя'
}
