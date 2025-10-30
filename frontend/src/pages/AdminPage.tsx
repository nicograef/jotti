import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export function AdminPage() {
  return (
    <Tabs defaultValue="produkte" className="min-h-screen">
      {/* Main content area; add bottom padding so it doesn't get covered by the bottom bar */}
      <div className="mx-auto w-full max-w-3xl px-4 pt-4 pb-24">
        <TabsContent value="produkte">
          Produkte verwalten …
        </TabsContent>
        <TabsContent value="tische">
          Tische verwalten …
        </TabsContent>
        <TabsContent value="benutzer">
          Benutzer verwalten …
        </TabsContent>
      </div>

      {/* Bottom-centered tab bar */}
      <div className="fixed inset-x-0 bottom-0 z-50 border-t bg-background/80 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto w-full max-w-3xl px-4 py-2">
          <TabsList className="grid w-full grid-cols-2 gap-2">
            <TabsTrigger className="w-full" value="produkte">Produkte</TabsTrigger>
            <TabsTrigger className="w-full" value="tische">Tische</TabsTrigger>
            <TabsTrigger className="w-full" value="benutzer">Benutzer</TabsTrigger>
          </TabsList>
        </div>
      </div>
    </Tabs>
  )
}
