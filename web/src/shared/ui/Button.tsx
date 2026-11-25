import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { clsx } from 'clsx'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  icon?: ReactNode
}

export function Button({
  className,
  variant = 'primary',
  icon,
  children,
  ...props
}: ButtonProps) {
  return (
    <button className={clsx('button', `button-${variant}`, className)} {...props}>
      {icon}
      <span>{children}</span>
    </button>
  )
}
