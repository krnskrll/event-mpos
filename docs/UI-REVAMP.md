# Bhoomy POS — UI revamp

Flutter app in `mobile/`. Mockups: [canvas](https://claude.ai/artifact/J8M57FS2H7E9Daj34bFZN6) (one page per phase).

## Decisions

| # | Decision | Date |
|---|---|---|
| U1 | Follow Bhoomy Brand Guidelines v2.0 as used by Bhoomy web (`cuscos-frontend/src/theme/index.js`), not Paradise/Paragon blue | 2026-11-02 |
| U2 | Softer corners than Bhoomy web (8–24 instead of 6–8): pressed with thumbs, not clicked | 2026-11-02 |
| U3 | App name **Bhoomy POS**; "Shade" → **Variant** (general retail) | 2026-11-02 |
| U4 | Warm, soft screens (inset photos, badges, shadows, floating panels); no emoji, glows or caption-on-everything | 2026-11-02 |
| U5 | One app, 3 size groups, 2 layouts: phone layout under 840 px, side-by-side tablet layout from 840 px | 2026-11-02 |
| U6 | Phone tabs Home / **Sell** / Orders; the Cart tab becomes a floating cart bar + full-screen order | 2026-11-02 |
| U7 | Card **+** adds one (opens the sheet if the product has variants); tapping the card opens the sheet. Save draft lives inside Drafts | 2026-11-02 |
| U8 | Offline keeps selling (chip + gold note only); **event closed** stops selling (Sell tab disabled, closed card). A status chip replaces the full-width banners | 2026-11-02 |
| U9 | End session: tablet Home header and the avatar menu (all sizes); Orders tab shows a badge for orders still sending | 2026-11-02 |
| U10 | Every order has a detail view (items, totals, progress) with a Details / Receipt switch; tablet shows it beside the list. Wording is "Sent to server", never the platform name | 2026-11-02 |
| U11 | Orders keep subtotal, discount and promotion name (database v3) so the detail shows what was charged | 2026-11-02 |

## Phases

| Phase | Scope | Mockup | Build |
|---|---|---|---|
| 1 Foundation | Tokens, Lato, type scale, component themes, shared widgets, wording | ✅ | ✅ PR #15 |
| 2 Selling | Tablet products + order panel, phone grid + cart bar, product sheet, cart | ✅ | ✅ PR #17 |
| 3 Shell + Home | Floating forest rail, phone bottom bar, Home dashboard, status in top bar | ✅ | ✅ PR #18 |
| 4 Orders + Receipt | Order list + detail side by side on tablet, receipt card | ✅ | ✅ PR #20 |
| 5 Entry flow | Login, events, sync | — | — |
| 6 Utilities | Printer, customer scan, scanner, drafts | — | — |

## Tokens (`mobile/lib/app/theme.dart` → `Brand`)

| Token | Hex | Use |
|---|---|---|
| `green` | `#2E7D53` | Main buttons, selected states, + |
| `forest` | `#143A24` | Headings, totals, tablet rail, cart bar |
| `gold` | `#D99A2B` | Promos, member tier, counts on dark (dark text only) |
| `canvas` / `surface` | `#F7F6F1` / `#FFFFFF` | App background / cards |
| `sunken` / `divider` / `cardLine` | `#EDF1F6` / `#DCE3EC` / `#E7E4DA` | Stepper track / hairlines / card border |
| `fieldLine` | `#8A94A0` | Input borders (3:1 contrast) |
| `ink` / `slate` / `muted` | `#33415A` / `#47566B` / `#64748B` | Body / secondary on canvas / meta on white |
| Status | success `#12885E`, warning `#C98A00`, error `#D64550`, info `#2F6FB5` | Each has `*Ink` (text) and `*Soft` (background) |

## Type (`AppText`, Lato bundled, OFL)

| Style | Size / weight | Use |
|---|---|---|
| `total` | 28 / 900 | Order and receipt total |
| `title` | 24 / 900 | Screen title |
| `heading` | 20 / 700 | Panel title |
| `subhead` | 18 / 700 | Product name on detail |
| `bodyLarge` | 16 / 700 | Prices, emphasis |
| `body` | 14 / 400 | Default text |
| `label` / `button` | 13 / 700, 15 / 700 | Chips, field labels / buttons |
| `caption` | 12 / 400 | Meta |

Every style names `fontFamily: 'Lato'`: button and chip text don't inherit the app font. Prices use tabular figures.

## Shape

| Radius | Use |
|---|---|
| 8 | Stepper buttons |
| 12–14 | Buttons, fields, thumbnails |
| 16–18 | Cards |
| 24 | Panels, sheets, dialogs, tablet rail |
| full | Chips, badges |

Touch targets ≥ 44 px. Main action 56 px.

## Shared widgets (`mobile/lib/app/`)

| Widget | File | Notes |
|---|---|---|
| `BrandButton` | `widgets.dart` | Full-width 56 px; `secondary` = outlined |
| `ScreenHeader` | `widgets.dart` | Light title + subtitle + actions |
| `QtyStepper` | `widgets.dart` | Grey track, white −, green +; null callback disables a side |
| `CountChip` | `widgets.dart` | Filter chip with its count in a pill |
| `StockBadge` / `stockStatus()` | `widgets.dart` | Badge on the photo: "Sold out" / "Only n left" / "n left" |
| `OrderPanel` | `features/cart/order_panel.dart` | Customer, lines, promo, total, Confirm; tablet side panel and phone cart page |
| `showProductSheet` | `features/sell/product_sheet.dart` | Bottom sheet on phone, dialog on tablet; card taps and scans |
| `Pill`, `SectionCard`, `EmptyState`, `ProductImage` | `widgets.dart` | |
| `BhoomyLogo` | `logo.dart` | Drawn in code from the brand SVG |
| `OrderDetail` / `OrderDetailScreen` | `features/orders/order_detail.dart` | Details / Receipt switch, items, totals, progress, Reprint, Cancel |
| `ReceiptPaper`, `reprintReceipt` | `features/receipt/receipt_paper.dart` | Centred 58 mm slip, aligned as printed |
| `SyncChip`, `UserAvatar`, `showSessionMenu` | `status.dart` | Synced / Sending n / Offline · n to send; avatar opens printer + End session |

## Open items

- [x] Warm `StockBadge`, `QtyStepper`, `CountChip` (Phase 2)
- [ ] **Parked:** customer is required to confirm (EPIC-04/07 rule; backend `customer.name` required). Options: walk-in default (app only), fully optional (DB + API), or per-outlet setting — decide with Bhoomy outlets
- [ ] Launch screen and launcher icon still Flutter default
- [x] Status-bar icons over the forest rail on tablet (rail now floats inside the safe area)

## Phase 2 checks (emulator, phone + tablet width, Go backend — 2026-11-02)

| # | Check | Result |
|---|---|---|
| 1 | Phone Sell: chips with counts, stock badges, "n in order", cart bar above 3 tabs | ✅ |
| 2 | Card + adds one; bar total and promo update (promo has a minimum subtotal) | ✅ |
| 3 | Card tap → product sheet above the tab bar → Add to order | ✅ |
| 4 | Review → order page: identify customer (typed name) → Confirm → receipt #002 → New order clears | ✅ |
| 5 | Tablet width: products + always-open order panel; product dialog adds to panel | ✅ |

## Phase 3 checks (emulator, phone + tablet width — 2026-11-02)

| # | Check | Result |
|---|---|---|
| 1 | Phone Home: greeting, Synced chip, Start selling (Scan, Drafts), 4 stats, recent orders, floating bottom bar | ✅ |
| 2 | Airplane mode: chip "Offline", gold "Keep selling" note, Sell still enabled | ✅ |
| 3 | Offline sale → #003 Sending, chip "Offline · 1 to send", Orders badge 1 | ✅ |
| 4 | Reconnect → #003 Confirmed, badge clears | ✅ |
| 5 | Tablet: floating rail (printer dot, avatar), End session, event + device + live promotions | ✅ |
| 6 | Clear order asks first, empties the order | ✅ |

## Phase 4 checks (emulator, phone + tablet width — 2026-11-02)

| # | Check | Result |
|---|---|---|
| 1 | Upgrade from database v2 keeps old orders; new order saves subtotal, discount, promotion name | ✅ |
| 2 | After Confirm: "Order #004 confirmed", receipt printed, "Sent to server", centred receipt, New order | ✅ |
| 3 | Orders: count chips, search, reason line per order, real item counts | ✅ |
| 4 | Phone detail: items, "10% off everything" discount, progress to "Sent to server", Reprint, Cancel | ✅ |
| 5 | Details / Receipt switch shows the centred slip | ✅ |
| 6 | Tablet: list beside the selected order's detail | ✅ |
