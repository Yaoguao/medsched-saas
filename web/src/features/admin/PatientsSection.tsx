import { useQuery } from '@tanstack/react-query'
import type { FormEvent } from 'react'
import { useMemo, useState } from 'react'
import { HeartPulse, Link2, Plus, Stethoscope, UsersRound } from 'lucide-react'
import type {
  CreatePatientPayload,
  LinkPatientUserPayload,
  Patient,
  UpdatePatientPayload,
  UpdateMedicalProfilePayload,
  User,
} from '../../shared/api/types'
import { coreApi } from '../../shared/api/core'
import { getPatientMedicalProfile } from '../../shared/lib/medicalProfile'
import { Button } from '../../shared/ui/Button'
import { MetricCard } from '../../shared/ui/MetricCard'
import { OperationFeedback } from '../../shared/ui/OperationFeedback'
import {
  PatientWorkbench,
  RiskFlags,
} from '../patients/PatientWorkbench'
import { patientRiskFlags } from '../patients/patientRisk'
import { AdminTable } from './AdminTable'
import type { ActionResult } from './types'
import {
  formBoolean,
  formDate,
  formNumber,
  formOptionalDate,
  formOptionalNumber,
  formOptionalString,
  formString,
} from './form'

type PatientsSectionProps = {
  patients: Patient[]
  users: User[]
  initialPatientId?: string
  actionResult?: ActionResult | null
  onDismissResult?: () => void
  isPending: boolean
  onCreatePatient: (payload: CreatePatientPayload) => void
  onUpdatePatient: (patientId: string, payload: UpdatePatientPayload) => void
  onUpdateProfile: (patientId: string, payload: UpdateMedicalProfilePayload) => void
  onLinkPatientUser: (patientId: string, payload: LinkPatientUserPayload) => void
  onOpenIntegrations: (patientId: string) => void
  onBookPatient: (patientId: string) => void
}

export function PatientsSection({
  patients,
  users,
  initialPatientId = '',
  actionResult = null,
  onDismissResult = () => undefined,
  isPending,
  onCreatePatient,
  onUpdatePatient,
  onUpdateProfile,
  onLinkPatientUser,
  onOpenIntegrations,
  onBookPatient,
}: PatientsSectionProps) {
  const [selectedPatientId, setSelectedPatientId] = useState(initialPatientId)
  const [query, setQuery] = useState('')
  const selectedPatientQuery = useQuery({
    queryKey: ['patient', selectedPatientId],
    queryFn: () => coreApi.patient(selectedPatientId),
    enabled: Boolean(selectedPatientId),
  })

  const filteredPatients = useMemo(() => {
    const search = query.trim().toLowerCase()
    if (!search) {
      return patients
    }
    return patients.filter((patient) =>
      [patient.full_name, patient.email, patient.phone, patient.neighbourhood, patient.id]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)),
    )
  }, [patients, query])

  function createPatient(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreatePatient({
      full_name: formString(form, 'full_name'),
      gender: formString(form, 'gender', 'UNKNOWN'),
      birth_date: formDate(form, 'birth_date'),
      age: formNumber(form, 'age'),
      phone: formString(form, 'phone'),
      email: formString(form, 'email'),
      neighbourhood: formString(form, 'neighbourhood'),
      scholarship: formBoolean(form, 'scholarship'),
      external_ehr_id: formOptionalString(form, 'external_ehr_id'),
      medical_profile: {
        chronic_conditions: formString(form, 'chronic_conditions'),
        risk_factors: formString(form, 'risk_factors'),
        allergies: formString(form, 'allergies'),
        notes: formString(form, 'notes'),
        has_hypertension: formBoolean(form, 'has_hypertension'),
        has_diabetes: formBoolean(form, 'has_diabetes'),
        has_alcoholism: formBoolean(form, 'has_alcoholism'),
        has_handicap: formBoolean(form, 'has_handicap'),
      },
    })
  }

  function updateProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onUpdateProfile(selectedPatientId, {
      chronic_conditions: formString(form, 'chronic_conditions'),
      risk_factors: formString(form, 'risk_factors'),
      allergies: formString(form, 'allergies'),
      notes: formString(form, 'notes'),
      has_hypertension: formBoolean(form, 'has_hypertension'),
      has_diabetes: formBoolean(form, 'has_diabetes'),
      has_alcoholism: formBoolean(form, 'has_alcoholism'),
      has_handicap: formBoolean(form, 'has_handicap'),
    })
  }

  function updatePatient(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onUpdatePatient(formString(form, 'patient_id', selectedPatientId), {
      full_name: formOptionalString(form, 'full_name'),
      gender: formOptionalString(form, 'gender'),
      birth_date: formOptionalDate(form, 'birth_date'),
      age: formOptionalNumber(form, 'age'),
      phone: formOptionalString(form, 'phone'),
      email: formOptionalString(form, 'email'),
      neighbourhood: formOptionalString(form, 'neighbourhood'),
      scholarship: form.get('scholarship_mode') === 'set' ? formBoolean(form, 'scholarship') : undefined,
      external_ehr_id: formOptionalString(form, 'external_ehr_id'),
    })
  }

  function linkPatientUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onLinkPatientUser(formString(form, 'patient_id', selectedPatientId), {
      user_id: formString(form, 'user_id'),
    })
  }

  const unlinkedPatients = patients.filter((patient) => !patient.user_id)
  const linkableUsers = users.filter((user) => user.role === 'PATIENT' && !user.patient_id)
  const linkedCount = patients.filter((patient) => patient.user_id).length
  const ehrSyncedCount = patients.filter((patient) => patient.last_ehr_sync_at || patient.external_ehr_id).length
  const riskCount = patients.filter((patient) => patientRiskFlags(patient).length > 0).length
  const selectedPatient = selectedPatientQuery.data ?? patients.find((patient) => patient.id === selectedPatientId)
  const selectedProfile = getPatientMedicalProfile(selectedPatient)
  const linkPatientTargetId = selectedPatient && !selectedPatient.user_id
    ? selectedPatient.id
    : ''

  return (
    <>
    <OperationFeedback result={actionResult} onDismiss={onDismissResult} />
    <PatientWorkbench
      patients={filteredPatients}
      selectedPatient={selectedPatient}
      selectedPatientId={selectedPatientId}
      isLoadingDetails={selectedPatientQuery.isFetching}
      isPending={isPending}
      onSelectPatient={setSelectedPatientId}
      onOpenIntegrations={onOpenIntegrations}
      onBookPatient={onBookPatient}
    />

    <section className="metrics-grid">
      <MetricCard label="Пациенты" value={patients.length} hint="в текущей выборке" tone="blue" icon={<UsersRound size={22} />} />
      <MetricCard label="Аккаунты" value={linkedCount} hint="привязаны к картам" tone="green" icon={<Link2 size={22} />} />
      <MetricCard label="EHR/EMR" value={ehrSyncedCount} hint="есть внешняя синхронизация" tone="violet" icon={<Stethoscope size={22} />} />
      <MetricCard label="Признаки риска" value={riskCount} hint="по медпрофилю" tone="orange" icon={<HeartPulse size={22} />} />
    </section>

    <section className="admin-two-column">
      <details className="panel workspace-form-disclosure">
        <summary>
          <Plus size={18} />
          <div>
            <p className="eyebrow">Пациенты</p>
            <strong>Добавить пациента</strong>
          </div>
        </summary>
        <form className="admin-form" onSubmit={createPatient}>
          <input name="full_name" placeholder="ФИО" required />
          <input name="gender" placeholder="Пол: M/F/UNKNOWN" defaultValue="UNKNOWN" />
          <input name="birth_date" type="date" defaultValue="1990-01-01" />
          <input name="age" type="number" min="0" placeholder="Возраст" defaultValue="0" />
          <input name="phone" placeholder="Телефон" />
          <input name="email" type="email" placeholder="Email" />
          <input name="neighbourhood" placeholder="Район" />
          <input name="external_ehr_id" placeholder="EHR ID, если уже известен" />
          <label className="checkbox-line">
            <input name="scholarship" type="checkbox" />
            Льгота
          </label>
          <input name="chronic_conditions" placeholder="Хронические заболевания" />
          <input name="risk_factors" placeholder="Факторы риска" />
          <input name="allergies" placeholder="Аллергии" />
          <input name="notes" placeholder="Заметки" />
          <div className="checkbox-grid">
            <label><input name="has_hypertension" type="checkbox" /> Гипертония</label>
            <label><input name="has_diabetes" type="checkbox" /> Диабет</label>
            <label><input name="has_alcoholism" type="checkbox" /> Алкоголизм</label>
            <label><input name="has_handicap" type="checkbox" /> Инвалидность</label>
          </div>
          <Button type="submit" disabled={isPending}>Создать пациента</Button>
        </form>
      </details>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Аккаунт</p>
            <h2>Привязать аккаунт к пациенту</h2>
          </div>
        </div>
        <form className="admin-form" onSubmit={linkPatientUser}>
          <select name="patient_id" value={linkPatientTargetId} onChange={(event) => setSelectedPatientId(event.target.value)} required>
            <option value="">Пациент без аккаунта</option>
            {unlinkedPatients.map((patient) => (
              <option key={patient.id} value={patient.id}>
                {patient.full_name} · {patient.email || patient.phone || patient.id}
              </option>
            ))}
          </select>
          <select name="user_id" required>
            <option value="">Аккаунт пациента без карты</option>
            {linkableUsers.map((user) => (
              <option key={user.id} value={user.id}>
                {user.email} · {user.phone || user.id}
              </option>
            ))}
          </select>
          <Button type="submit" disabled={isPending || !linkPatientTargetId}>
            Привязать аккаунт
          </Button>
        </form>
        <p className="section-note">
          Аккаунт пациента создается отдельно. Карта пациента и медпрофиль остаются под контролем регистратора.
        </p>
      </article>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Пациент</p>
            <h2>Редактировать данные</h2>
          </div>
        </div>
        <form className="admin-form" onSubmit={updatePatient} key={selectedPatient?.id ?? 'patient-edit-empty'}>
          <select name="patient_id" value={selectedPatientId} onChange={(event) => setSelectedPatientId(event.target.value)} required>
            <option value="">Выбери пациента</option>
            {patients.map((patient) => (
              <option key={patient.id} value={patient.id}>
                {patient.full_name} · {patient.email || patient.phone || patient.id}
              </option>
            ))}
          </select>
          <input name="full_name" defaultValue={selectedPatient?.full_name ?? ''} placeholder="ФИО" />
          <select name="gender" defaultValue={selectedPatient?.gender ?? 'UNKNOWN'}>
            <option value="UNKNOWN">Пол не указан</option>
            <option value="M">Мужской</option>
            <option value="F">Женский</option>
          </select>
          <input name="birth_date" type="date" defaultValue={dateInputValue(selectedPatient?.birth_date)} />
          <input name="age" type="number" min="0" defaultValue={selectedPatient?.age ?? 0} placeholder="Возраст" />
          <input name="phone" defaultValue={selectedPatient?.phone ?? ''} placeholder="Телефон" />
          <input name="email" type="email" defaultValue={selectedPatient?.email ?? ''} placeholder="Email" />
          <input name="neighbourhood" defaultValue={selectedPatient?.neighbourhood ?? ''} placeholder="Район" />
          <input name="external_ehr_id" defaultValue={selectedPatient?.external_ehr_id ?? ''} placeholder="EHR ID" />
          <label className="checkbox-line">
            <input name="scholarship_mode" type="checkbox" value="set" defaultChecked={Boolean(selectedPatient)} />
            Изменить льготу
          </label>
          <label className="checkbox-line">
            <input name="scholarship" type="checkbox" defaultChecked={selectedPatient?.scholarship ?? false} />
            Льгота включена
          </label>
          <Button type="submit" disabled={isPending || !selectedPatientId}>Сохранить пациента</Button>
        </form>
      </article>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Медпрофиль</p>
            <h2>Обновить признаки</h2>
          </div>
        </div>
        <form className="admin-form" onSubmit={updateProfile} key={selectedPatient?.id ?? 'profile-edit-empty'}>
          <select
            value={selectedPatientId}
            onChange={(event) => setSelectedPatientId(event.target.value)}
            required
          >
            <option value="">Выбери пациента</option>
            {patients.map((patient) => (
              <option key={patient.id} value={patient.id}>
                {patient.full_name} · {patient.email || patient.phone || patient.id}
              </option>
            ))}
          </select>
          <input name="chronic_conditions" defaultValue={selectedProfile?.chronic_conditions ?? ''} placeholder="Хронические заболевания" />
          <input name="risk_factors" defaultValue={selectedProfile?.risk_factors ?? ''} placeholder="Факторы риска" />
          <input name="allergies" defaultValue={selectedProfile?.allergies ?? ''} placeholder="Аллергии" />
          <input name="notes" defaultValue={selectedProfile?.notes ?? ''} placeholder="Заметки" />
          <div className="checkbox-grid">
            <label><input name="has_hypertension" type="checkbox" defaultChecked={selectedProfile?.has_hypertension ?? false} /> Гипертония</label>
            <label><input name="has_diabetes" type="checkbox" defaultChecked={selectedProfile?.has_diabetes ?? false} /> Диабет</label>
            <label><input name="has_alcoholism" type="checkbox" defaultChecked={selectedProfile?.has_alcoholism ?? false} /> Алкоголизм</label>
            <label><input name="has_handicap" type="checkbox" defaultChecked={selectedProfile?.has_handicap ?? false} /> Инвалидность</label>
          </div>
          <Button type="submit" disabled={isPending || !selectedPatientId}>
            Обновить профиль
          </Button>
        </form>
      </article>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Список</p>
            <h2>Пациенты</h2>
          </div>
          <input
            className="table-filter"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Поиск по ФИО, телефону, email"
          />
        </div>
        <AdminTable
          items={filteredPatients}
          empty="Пациентов пока нет."
          columns={[
            { key: 'name', title: 'ФИО', render: (patient) => patient.full_name },
            { key: 'age', title: 'Возраст', render: (patient) => patient.age },
            { key: 'phone', title: 'Телефон', render: (patient) => patient.phone },
            { key: 'risk', title: 'Мед. признаки', render: (patient) => <RiskFlags patient={patient} /> },
            { key: 'ehr', title: 'EHR', render: (patient) => (
              <span className={`workspace-badge ${patient.external_ehr_id ? 'info' : 'neutral'}`}>
                {patient.external_ehr_id ? 'Связан' : 'Не синхр.'}
              </span>
            ) },
            { key: 'user', title: 'Аккаунт', render: (patient) => patient.user_id ? <span className="workspace-badge success">Привязан</span> : <span className="workspace-badge neutral">Не привязан</span> },
            { key: 'actions', title: 'Действия', render: (patient) => (
              <Button type="button" variant="secondary" onClick={() => setSelectedPatientId(patient.id)}>
                Открыть
              </Button>
            ) },
          ]}
        />
      </article>
    </section>
    </>
  )
}

function dateInputValue(value?: string) {
  if (!value) {
    return ''
  }
  return value.slice(0, 10)
}
