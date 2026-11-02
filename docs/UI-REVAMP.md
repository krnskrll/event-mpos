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

## Phases

| Phase | Scope | Mockup | Build |
|---|---|---|---|
| 1 Foundation | Tokens, Lato, type scale, component themes, shared widgets, wording | ✅ | ✅ PR #15 |
| 2 Selling | Tablet products + order panel, phone grid + cart bar, product sheet, cart | ✅ in review | — |
| 3 Shell + Home | Floating forest rail, phone bottom bar, Home dashboard, status in top bar | — | — |
| 4 Orders + Receipt | Order list + detail side by side on tablet, receipt card | — | — |
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
| `QtyStepper` | `widgets.dart` | − qty +; null callback disables a side |
| `StockLabel` / `stockStatus()` | `widgets.dart` | "Sold out" / "Only n left" / "n left" |
| `Pill`, `SectionCard`, `EmptyState`, `ProductImage` | `widgets.dart` | |
| `BhoomyLogo` | `logo.dart` | Drawn in code from the brand SVG |

## Open items

- [ ] Phase 2 build restyles `StockLabel` → badge on photo, `QtyStepper` → grey track + solid green +, chips → count pill (warm style, U4)
- [ ] Launch screen and launcher icon still Flutter default
- [ ] Status-bar icons over the forest rail on tablet (Phase 3)
