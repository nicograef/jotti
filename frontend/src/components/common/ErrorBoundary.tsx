import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
}

/** ErrorBoundary component to catch JavaScript errors in child components
 *
 * React error boundaries must be class components.
 * React does not provide a hook equivalent for componentDidCatch or getDerivedStateFromError
 * — these lifecycle methods are only available on class components.
 * There is no useErrorBoundary hook in React core.
 *
 * This is a known React limitation that has persisted through React 18 and 19.
 * It's the one remaining case where a class component is required.
 */
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen items-center justify-center bg-background p-4 text-foreground">
          <div className="max-w-md text-center">
            <h1 className="mb-4 text-2xl font-bold">
              Etwas ist schiefgelaufen
            </h1>
            <p className="mb-6 text-muted-foreground">
              Ein unerwarteter Fehler ist aufgetreten. Bitte laden Sie die Seite
              neu.
            </p>
            <button
              onClick={() => {
                window.location.reload()
              }}
              className="rounded-md bg-primary px-6 py-2 text-primary-foreground hover:bg-primary/90"
            >
              Neu laden
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
