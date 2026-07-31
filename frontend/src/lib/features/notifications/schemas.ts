/**
 * Schemas Zod de los formularios de notificaciones.
 *
 * Vivió un tiempo en `features/settings` —escribe el mismo endpoint de
 * preferencias del usuario— pero es de esta pantalla, y aquí es donde se busca.
 */

import { z } from 'zod';

export const notificationPreferencesSchema = z.object({
	emailAlerts: z.coerce.boolean(),
	weeklySummary: z.coerce.boolean()
});
