# El blog

El blog público vive en `finexia.me/blog`. Los artículos son archivos Markdown
del repositorio: se convierten en HTML durante el build y se sirven
prerenderizados. **Publicar un artículo es añadir un `.md` y desplegar.**

No hay CMS, no hay base de datos y no hay panel de administración. El historial
de cada artículo es el de git, y una errata se corrige en un pull request como
cualquier otro cambio.

## Escribir un artículo

Crea `frontend/src/content/blog/<slug>.md`. El nombre del archivo es la URL:
`por-que-existe-finexia.md` se publica en `/blog/por-que-existe-finexia`.

```markdown
---
title: Por qué existe Finexia
description: Tu dinero está repartido en cinco sitios y ninguno te enseña el total.
date: 2026-09-15
tags: [producto]
---

El cuerpo, en Markdown normal. Tablas y listas incluidas (GFM).
```

### La cabecera

| Campo         | Obligatorio | Qué es                                                                               |
| ------------- | ----------- | ------------------------------------------------------------------------------------ |
| `title`       | Sí          | El titular. Sale en el índice, en el `<title>` y en el Open Graph.                   |
| `description` | Sí          | Una o dos frases. Es el resumen del índice, la `meta description` y el item del RSS. |
| `date`        | Sí          | `AAAA-MM-DD`. Ordena el índice y es la `datePublished` del JSON-LD.                  |
| `tags`        | Sí          | Al menos una. Cada etiqueta genera su página en `/blog/tag/<etiqueta>`.              |
| `author`      | No          | Por defecto, `Equipo Finexia`.                                                       |
| `draft`       | No          | `true` se ve en `pnpm dev` y no se publica.                                          |

La validan `schemas.ts` con Zod y `frontmatter.ts`, que lee un subconjunto
pequeño de YAML: pares `clave: valor`, comillas, booleanos y listas en línea.
Si falta un campo, **el build falla** nombrando el archivo y el campo.

### Los borradores

Con `draft: true` el artículo se ve en `pnpm dev`, marcado como **Borrador** en
el índice y en su página, y el build de producción lo ignora: no se
prerenderiza, ni entra en el índice, las etiquetas, el RSS o el sitemap. Su URL
da 404 en producción. Publicarlo es quitar la línea (o poner `draft: false`) y
desplegar.

Una etiqueta que solo usan borradores tampoco se publica.

### Las etiquetas

Se escriben tal cual se leen —`guías`, con tilde— y la URL se deriva sin ella
(`/blog/tag/guias`). `Guías` y `guias` son la misma etiqueta. No hace falta
declararlas en ningún sitio: la lista del blog sale de las que usan los
artículos.

Una etiqueta que se queda sin artículos deja de tener página, y eso es lo
correcto: su URL pasa a devolver 404 en vez de una página vacía.

### El tono

La voz es la de la portada: frases cortas, segunda persona, nada de jerga y
nada que el producto no haga todavía. Si un artículo promete algo que el panel
no tiene, sobra la promesa, no el artículo.

## Comprobar antes de publicar

```sh
cd frontend
pnpm dev             # /blog, el artículo y su etiqueta
pnpm lint            # prettier también formatea src/content/**
pnpm test:unit -- --run
pnpm build           # el prerender falla si una cabecera está mal
```

`posts.spec.ts` lee los artículos de verdad, no una fixture: una cabecera mal
escrita rompe los tests antes de romper el build.

## Qué se genera solo

Al añadir un artículo, sin tocar nada más:

- entra en `/blog`, en la página de cada una de sus etiquetas y en el feed
  `/blog/rss.xml`;
- entra en `/sitemap.xml` con su fecha real, y sus etiquetas nuevas también;
- se prerenderiza como HTML con su canonical, su Open Graph y su JSON-LD
  `BlogPosting`.

## Cómo está montado

| Dónde                              | Qué                                                              |
| ---------------------------------- | ---------------------------------------------------------------- |
| `frontend/src/content/blog/*.md`   | Los artículos.                                                   |
| `frontend/src/lib/features/blog/`  | La feature: carga, helpers puros, schemas y componentes.         |
| `frontend/src/routes/blog/`        | Índice, artículo, etiqueta y feed. Prerenderizados salvo el RSS. |
| `frontend/src/routes/sitemap.xml/` | Lee los artículos de la feature.                                 |

`posts.ts` carga los `.md` con `import.meta.glob` **sin `eager`** y hace
`await import('marked')`. Así el conversor de Markdown y el texto de los
artículos se quedan en el servidor y no engordan el bundle del navegador. Si
alguien lo pone en `eager` o importa `marked` arriba del archivo nada falla,
solo crece el bundle: lo vigila el test e2e «el cliente no descarga el conversor
de Markdown».

El cuerpo se pinta con `{@html}` sin sanear, porque viene de un archivo del
repositorio que pasó por revisión. Si algún día el contenido viniera de fuera,
habría que sanearlo antes de pintarlo.
