/**
 * Utility function to conditionally combine classNames
 */
export function cn(...classes: (string | Record<string, boolean> | undefined | null)[]): string {
	return classes
		.flatMap((cls) => {
			if (!cls) return [];
			if (typeof cls === 'string') return cls;
			return Object.entries(cls)
				.filter(([, value]) => value)
				.map(([key]) => key);
		})
		.filter(Boolean)
		.join(' ');
}
