import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'expo-router'
import { useState } from 'react'
import { useForm, useController } from 'react-hook-form'

import { DevLoginScreen } from '@/features/auth/components/presentational/dev-login-screen'
import { useDevLogin } from '@/features/auth/hooks/use-dev-login'
import { devLoginFormSchema, type DevLoginFormValues } from '@/features/auth/types/dev-login.types'
import {
  getRecentDevAccounts,
  type RecentDevAccount,
} from '@/features/auth/lib/dev-login-recent-accounts'

export function DevLoginContainer() {
  const router = useRouter()
  const { signIn, isLoading, error } = useDevLogin()
  // Read once on mount - this screen's own successful login navigates away
  // (app/dev-login.tsx redirects once session is set), so the list never
  // needs to react to further writes within the same mount.
  const [recentAccounts] = useState<RecentDevAccount[]>(() => getRecentDevAccounts())

  const {
    control,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<DevLoginFormValues>({
    resolver: zodResolver(devLoginFormSchema),
    defaultValues: { name: '', admin: false },
  })
  const {
    field: { value: name, onChange: onNameChange },
  } = useController({ name: 'name', control })
  const {
    field: { value: admin, onChange: onAdminChange },
  } = useController({ name: 'admin', control })

  const onSubmit = handleSubmit(({ name: submittedName, admin: submittedAdmin }) => {
    void signIn(submittedName, submittedAdmin)
  })

  const onPickRecentAccount = (account: RecentDevAccount) => {
    setValue('name', account.name)
    setValue('admin', account.admin)
    void signIn(account.name, account.admin)
  }

  return (
    <DevLoginScreen
      name={name}
      onNameChange={onNameChange}
      nameError={errors.name?.message ?? null}
      admin={admin}
      onAdminChange={onAdminChange}
      onSubmit={onSubmit}
      isLoading={isLoading}
      error={error}
      recentAccounts={recentAccounts}
      onPickRecentAccount={onPickRecentAccount}
      onBackToGoogle={() => router.replace('/login')}
    />
  )
}
