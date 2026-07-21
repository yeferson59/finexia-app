# `lib/shared` — utilidades transversales sin dominio

Código compartido que no pertenece a ninguna feature ni al design system.
Sustituye al antiguo cajón de sastre `lib/utils.ts`, repartido por tema.

## Contenido

- `css.ts` — `cn()`, combinación condicional de clases.
- `format/money.ts` — `formatCurrency()`.
- `format/date.ts` — `formatCalendarDate()`, `todayLocalDateString()`.
- `config/features.ts` — feature flags (`investments`, `selfRegistration`).
- (Fase 5) `privacy.svelte.ts` — estado transversal de privacidad.

`lib/utils.ts` sigue existiendo como re-export temporal de estos módulos para no
tocar a todos los importadores en la Fase 1; se elimina cuando ya nadie importe
`$lib/utils`.

## Reglas de dependencia

- `lib/shared` **no** importa de `lib/features`, `lib/api` ni `lib/ui`: es la
  capa más baja y todos pueden depender de ella.
