/**
 * La librería del botón de pagos de Bold, cargada solo cuando alguien pulsa
 * «Aportar»: quien solo lee la página no descarga nada de Bold.
 *
 * Integración personalizada (developers.bold.co): el script deja
 * `window.BoldCheckout`, que recibe la configuración firmada por el servidor
 * y con `open()` lleva a la pasarela.
 */

const SCRIPT_URL = 'https://checkout.bold.co/library/boldPaymentButton.js';

/** La configuración que firma `$lib/server/bold` y que Bold espera tal cual. */
export type BoldCheckoutOptions = Record<string, string>;

interface BoldCheckoutInstance {
	open(): void;
}

type BoldCheckoutConstructor = new (options: BoldCheckoutOptions) => BoldCheckoutInstance;

declare global {
	interface Window {
		BoldCheckout?: BoldCheckoutConstructor;
	}
}

let loading: Promise<BoldCheckoutConstructor> | null = null;

export function loadBoldCheckout(): Promise<BoldCheckoutConstructor> {
	if (window.BoldCheckout) return Promise.resolve(window.BoldCheckout);

	loading ??= new Promise<BoldCheckoutConstructor>((resolve, reject) => {
		const script = document.createElement('script');
		script.src = SCRIPT_URL;
		script.async = true;
		script.onload = () =>
			window.BoldCheckout ? resolve(window.BoldCheckout) : reject(new Error('BoldCheckout'));
		script.onerror = () => reject(new Error('BoldCheckout'));
		document.head.append(script);
	}).catch((error: unknown) => {
		// Un fallo de red no deja la promesa envenenada: el siguiente clic reintenta.
		loading = null;
		document.querySelector(`script[src="${SCRIPT_URL}"]`)?.remove();
		throw error;
	});

	return loading;
}

/** Abre la pasarela con una orden ya firmada. */
export async function openBoldCheckout(options: BoldCheckoutOptions): Promise<void> {
	const BoldCheckout = await loadBoldCheckout();
	new BoldCheckout(options).open();
}
