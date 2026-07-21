import { env } from '$env/dynamic/public';

/**
 * A flag is on only when its env var is exactly the string `"true"`. Anything
 * else (unset, `"false"`, `"0"`, …) leaves the feature off, so features stay
 * disabled by default until they are explicitly enabled.
 */
function isEnabled(value: string | undefined): boolean {
	return value === 'true';
}

/**
 * Application feature flags.
 *
 * `investments` gates the whole "Inversiones" section — the sidebar entry plus
 * every `/dashboard/investments` route. It ships disabled because the feature
 * is not part of the initial launch and is still being tested; enable it by
 * setting `PUBLIC_FEATURE_INVESTMENTS=true` in the environment.
 *
 * `selfRegistration` gates the public "Crear cuenta" form. It ships disabled
 * because Finexia is invite-only during the beta; enable it by setting
 * `PUBLIC_FEATURE_SELF_REGISTRATION=true` once open sign-up launches. Keep
 * this in sync with the backend's `SELF_REGISTRATION_ENABLED` — the API
 * rejects register requests independently, so this flag only controls the
 * frontend's UI/UX around it.
 */
export const features = {
	investments: isEnabled(env.PUBLIC_FEATURE_INVESTMENTS),
	selfRegistration: isEnabled(env.PUBLIC_FEATURE_SELF_REGISTRATION)
} as const;
