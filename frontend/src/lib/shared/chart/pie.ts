/**
 * Geometría del gráfico de torta: dónde cae cada porción en el lienzo.
 *
 * Es matemática de SVG y nada más —ni dominio, ni Svelte—, y por eso vive aquí
 * abajo: la usan el donut de asignación del panel (por clase de activo) y el de
 * concentración de la cartera (por activo), que son dos features distintas y no
 * pueden importarse entre sí. La copia que habría hecho falta si no bajara se
 * habría desviado en el primer ajuste de radio.
 *
 * El lienzo es fijo —200×200, radio 75— porque las dos gráficas lo comparten y
 * parametrizarlo solo serviría para que dejaran de parecerse.
 */

/** Centro y radio del lienzo. El `viewBox` que los acompaña es `0 0 200 200`. */
export const PIE_CENTER = 100;
export const PIE_RADIUS = 75;

/** Punto del círculo en grados, con 0º arriba (de ahí el -90). */
export function polarToCartesian(angle: number, radius: number, cx = PIE_CENTER, cy = PIE_CENTER) {
	const radians = (angle - 90) * (Math.PI / 180);
	return {
		x: cx + radius * Math.cos(radians),
		y: cy + radius * Math.sin(radians)
	};
}

/** Porción del donut como `path`, con el flag de arco largo por encima de 180º. */
export function generatePieSlice(
	percent: number,
	startAngle: number
): { d: string; startAngle: number; endAngle: number } {
	const cx = PIE_CENTER;
	const cy = PIE_CENTER;
	const radius = PIE_RADIUS;
	const endAngle = startAngle + (percent / 100) * 360;
	const largeArc = endAngle - startAngle > 180 ? 1 : 0;

	const startPoint = polarToCartesian(startAngle, radius, cx, cy);
	const endPoint = polarToCartesian(endAngle, radius, cx, cy);

	const d = [
		`M ${cx} ${cy}`,
		`L ${startPoint.x} ${startPoint.y}`,
		`A ${radius} ${radius} 0 ${largeArc} 1 ${endPoint.x} ${endPoint.y}`,
		'Z'
	].join(' ');

	return { d, startAngle, endAngle };
}

/**
 * Encadena las porciones: cada una arranca donde acabó la anterior.
 *
 * Genérica sobre lo que traiga cada gráfica —solo lee `percent`— para que la
 * entrada siga siendo la fila del dominio, con su etiqueta y su color, y no
 * haya que traducirla de ida y vuelta.
 */
export function buildSlices<T extends { percent: number }>(
	items: T[]
): (T & { d: string; startAngle: number; endAngle: number })[] {
	let angle = 0;
	return items.map((item) => {
		const slice = generatePieSlice(item.percent, angle);
		angle = slice.endAngle;
		return { ...item, ...slice };
	});
}
