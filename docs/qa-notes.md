# Notes from QA Session

### Internal Server Error for plausible domain error

**What?**
When a user sets up jotti freshly and tries to open a new (first) Kassensitzung on frontend/src/admin/kasse/KassensitzungPage.tsx the backend responds with 500 and the user only sees a general error message in the toast.

```sh
jotti-backend-local        | 9:13AM ERR Failed to check betreiber configuration error="not found" correlation=7bc17345
jotti-backend-local        | 9:13AM INF Request completed correlation=7bc17345 duration_ms=6 path=/admin/kassensitzung-eroeffnen status=500
jotti-reverse-proxy-local  | 172.18.0.1 - - [12/Jun/2026:09:13:09 +0000] "POST /api/admin/kassensitzung-eroeffnen HTTP/2.0" 500 33 "https://172.18.0.4/admin/kasse" "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0" "-"
```

**Suggestion**
If the Betreiber-Stammdaten are not found (i.e. not configured yet by user) this should cause a 4xx error with a proper error code so that the frontend can show a clear error message and guide the user what to do.

---

### toast has only neutral color

**What?**
The toast always appears in a neutral color/theme. The user cannot differentiate a successfull toast from a warning or error feedback.

**Suggestion**
Use info, success, warning and error themese for the toast given the context and semantics of the message.

---

### Produktvarianten ui glitch

**What?**
When activating or deactivating a product variant, the order of the product variants in the ui changes unintentionally.
See frontend/src/admin/products/AdminProductsPage.tsx

**Suggestion**
Stable sorting/ordering by id or name or created date.

---

### Direktverkauf is too prominent in the UI

**What?**
There are two modes in jotti: Table-Services and Direktverkauf. Some Vereine use only one mode for their Vereinsfeste. And the ones that use both dont need to switch easily: the users are either Servicekräfte (waiters) that take orders and payments from tables (mostly their selected favorites) OR are working at a stationary POS (Direktverkauf) where the is no service but customers order and pay at the cash register (Verkaufsstelle) and get a pick-up bon (Abholbon). This means one must only open the Direktverkauf Mode once at the beginning and then is working inside this mode for hours probably. And the service servicing tables dont use this mode at all.
See frontend/src/service/TableSelectionPage.tsx

**Suggestion**
Reconsider the Direkverkauf mode, the separation between the two modes and rework the UX and UI of jotti respectively.

---

### Refresh on favorite change

**What?**
When a service user selects or deselects a table as a favorite, the favorite list (Meine Tische) is not being refreshed (invalidate query) so the favorite os only shown (or hidden) on a hard page refresh.
See frontend/src/service/TableSelectionPage.tsx
See frontend/src/service/components/TischAuswahlDrawer.tsx

**Suggestion**
Use the invalidate query feature from tanstack query as used in similar other places (validate pattern in codebase).
See

---

### Non-Sticky primary action buttons

**What?**
On the table page for the tabs "Bestellen" and "Kassieren" the primary action buttons (i.e. submit/review Bestellung and submit/review Zahlung) are at the top of the page. However, the page is very long and users must scroll down to add products. In order to actually perform order or payment, service users must scroll to the top first which is cumbersome and bad ux.
See frontend/src/service/components/table/BestellungDrawer.tsx
See frontend/src/service/components/table/ZahlungDrawer.tsx

**Suggestion**
Rethink the UI and UX of the Table Page, evaluate floating action buttons, bottom buttons, sticky headers or sticky buttons.

---
