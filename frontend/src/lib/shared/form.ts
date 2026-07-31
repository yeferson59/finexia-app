/**
 * Reparto del `form` de una página entre secciones.
 *
 * Una página con varias form actions tiene un solo `form`: cada acción devuelve
 * `{ action: '<nombre>', … }` y cada sección tiene que decidir si ese resultado
 * es suyo. Estos helpers nacieron en `features/settings`; al aparecer el mismo
 * patrón en `features/notifications` bajaron aquí, que es donde vive lo que no
 * pertenece a ningún dominio.
 */

/** `form` de una página, sin tipar por acción. */
export type ActionForm = Record<string, unknown> | null | undefined;

/** `true` si `form` es el resultado correcto de esa acción. */
export function actionSucceeded(form: ActionForm, action: string): boolean {
	return form?.action === action && form?.success === true;
}

/** Mensaje de error de esa acción, o cadena vacía si el `form` no es suyo. */
export function actionError(form: ActionForm, action: string): string {
	return form?.action === action ? ((form?.error as string) ?? '') : '';
}

/**
 * Campo del resultado de una acción concreta, solo si esa acción tuvo éxito.
 * Sirve para los datos que devuelven las acciones (`imageUrl`, `secret`,
 * `recoveryCodes`…) sin volver a comprobar `action` + `success` en cada sitio.
 */
export function actionData<T>(form: ActionForm, action: string, field: string): T | undefined {
	return actionSucceeded(form, action) ? (form?.[field] as T | undefined) : undefined;
}
