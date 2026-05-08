import { env } from '$env/dynamic/private';
import { z } from 'zod';

const EnvSchema = z.object({
	baseApi: z.string()
});

export const envConfig = EnvSchema.parse({
	baseApi: env.BASE_API
});
