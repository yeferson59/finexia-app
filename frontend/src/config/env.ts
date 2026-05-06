import { BASE_API } from '$env/static/private';
import { z } from 'zod';

const EnvSchema = z.object({
	baseApi: z.string()
});

export const env = EnvSchema.parse({
	baseApi: BASE_API
});
