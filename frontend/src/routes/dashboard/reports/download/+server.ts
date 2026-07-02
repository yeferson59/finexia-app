import type { RequestHandler } from './$types';
import { authedFetchSafe } from '$lib/server/api';

const REPORTS: Record<string, { path: string; filename: string }> = {
	summary: { path: '/portfolios/export/summary', filename: 'resumen-mensual.xlsx' },
	transactions: { path: '/portfolios/export/transactions', filename: 'transacciones.xlsx' },
	risk: { path: '/portfolios/export/risk', filename: 'riesgo-volatilidad.xlsx' }
};

export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const type = url.searchParams.get('type') ?? '';
	const report = REPORTS[type];
	if (!report) return new Response('Not found', { status: 404 });

	const res = await authedFetchSafe({ cookies, fetch }, report.path);

	if (!res?.ok) return new Response('Error al generar el reporte', { status: res?.status ?? 502 });

	return new Response(res.body, {
		headers: {
			'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
			'Content-Disposition': `attachment; filename="${report.filename}"`
		}
	});
};
