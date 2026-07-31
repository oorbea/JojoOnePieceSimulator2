import { Component, type ReactNode } from 'react'

import { toAppError, type AppError } from '@/shared/api/errors'
import { ErrorFallback } from '@/shared/components/presentational/error-fallback'

type Props = { children: ReactNode }
type State = { error: AppError | null }

// Error boundaries must be class components — getDerivedStateFromError and
// componentDidCatch have no function-component equivalent.
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: unknown): State {
    return { error: toAppError(error) }
  }

  componentDidCatch(error: unknown, info: { componentStack: string }) {
    console.error('Unhandled render error', error, info.componentStack)
  }

  reset = () => this.setState({ error: null })

  render() {
    if (this.state.error) {
      return <ErrorFallback error={this.state.error} onRetry={this.reset} />
    }
    return this.props.children
  }
}
