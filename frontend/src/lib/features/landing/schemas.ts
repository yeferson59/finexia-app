/**
 * Schemas Zod de los formularios de la landing.
 *
 * El alta en la waitlist no es una form action sino un endpoint
 * (`routes/api/waitlist`), para que la landing pueda prerenderizarse como HTML
 * estático; el schema vive igualmente aquí, junto al formulario que lo envía.
 */

import { z } from 'zod';

export const waitlistEmailSchema = z.email();
