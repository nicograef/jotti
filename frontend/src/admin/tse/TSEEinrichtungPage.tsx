import { TSEEinrichtungWizard } from './TSEEinrichtungWizard'
import { TSEKonfigurationSection } from './TSEKonfigurationSection'

export function TSEEinrichtungPage() {
  return (
    <div className="flex flex-col gap-6 max-w-2xl">
      <TSEEinrichtungWizard />
      <TSEKonfigurationSection />
    </div>
  )
}
