import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';

const REPORTS: Record<string, { path: string; filename: string }> = {
	summary: { path: '/portfolios/export/summary', filename: 'resumen-mensual.xlsx' },
	transactions: { path: '/portfolios/export/transactions', filename: 'transacciones.xlsx' },
	risk: { path: '/portfolios/export/risk', filename: 'riesgo-volatilidad.xlsx' }
};

export const GET: RequestHandler = async ({ url, cookies }) => {
	const type = url.searchParams.get('type') ?? '';
	const report = REPORTS[type];
	if (!report) return new Response('Not found', { status: 404 });

	const accessToken = cookies.get('access_token_finexia');
	if (!accessToken) return new Response('Unauthorized', { status: 401 });

	const res = await fetch(`${env.BASE_API}${report.path}`, {
		headers: { Authorization: `Bearer ${accessToken}` }
	}).catch(() => null);

	if (!res?.ok) return new Response('Error al generar el reporte', { status: res?.status ?? 502 });

	return new Response(res.body, {
		headers: {
			'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
			'Content-Disposition': `attachment; filename="${report.filename}"`
		}
	});
};
