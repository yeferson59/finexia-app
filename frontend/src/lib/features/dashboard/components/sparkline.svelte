<script lang="ts">
	/*
	 * La curva del patrimonio, en miniatura y junto a la cifra.
	 *
	 * Contesta «¿y va subiendo?» sin bajar a la gráfica grande, que es la
	 * segunda pregunta de quien abre el panel y antes obligaba a hacer scroll.
	 *
	 * Es `aria-hidden` a propósito: los mismos datos están completos —con fechas
	 * y las dos series— en la tabla del lector de pantalla que acompaña a la
	 * gráfica de más abajo, en esta misma página. Repetirlos aquí sería obligar
	 * a recorrer catorce meses de puntos dos veces.
	 */
	interface Props {
		values: number[];
		width?: number;
		height?: number;
	}

	let { values, width = 260, height = 54 }: Props = $props();

	const PAD = 6;

	const path = $derived.by(() => {
		if (values.length < 2) return { line: '', area: '', endX: 0, endY: 0 };

		const min = Math.min(...values);
		const max = Math.max(...values);
		// Serie plana: sin ventana propia todos los puntos caerían en el borde
		// superior y la línea se leería como una subida al máximo.
		const span = max - min || Math.abs(max) || 1;

		const x = (i: number) => PAD + (i / (values.length - 1)) * (width - PAD * 2);
		const y = (v: number) => height - PAD - ((v - min) / span) * (height - PAD * 2);

		const line = values.map((v, i) => `${x(i).toFixed(1)},${y(v).toFixed(1)}`).join(' ');
		const last = values.length - 1;

		return {
			line,
			area: `${line} ${x(last).toFixed(1)},${height} ${x(0).toFixed(1)},${height}`,
			endX: x(last),
			endY: y(values[last])
		};
	});
</script>

{#if path.line}
	<svg
		class="spark"
		viewBox="0 0 {width} {height}"
		preserveAspectRatio="none"
		aria-hidden="true"
		focusable="false"
	>
		<defs>
			<linearGradient id="sparkFade" x1="0" y1="0" x2="0" y2="1">
				<stop offset="0%" stop-color="var(--amber)" stop-opacity="0.18" />
				<stop offset="100%" stop-color="var(--amber)" stop-opacity="0" />
			</linearGradient>
		</defs>
		<polygon points={path.area} fill="url(#sparkFade)" />
		<polyline points={path.line} />
		<!--
			El punto final es un segmento de longitud cero con el extremo redondo, no
			un `<circle>`: el lienzo se estira solo en horizontal —`preserveAspectRatio`
			es `none` para que la curva llene el ancho—, y un círculo salía ovalado y
			recortado contra el borde derecho. Un trazo que no escala sale redondo
			mida lo que mida la caja.
		-->
		<line x1={path.endX} y1={path.endY} x2={path.endX} y2={path.endY} class="end" />
	</svg>
{/if}

<style>
	.spark {
		display: block;
		width: 100%;
		height: 54px;
	}

	polyline {
		fill: none;
		stroke: var(--amber);
		stroke-width: 2;
		stroke-linecap: round;
		stroke-linejoin: round;
		/* El viewBox se estira en horizontal; sin esto el trazo se deforma con él. */
		vector-effect: non-scaling-stroke;
	}

	.end {
		stroke: var(--amber-light);
		stroke-width: 7;
		stroke-linecap: round;
		vector-effect: non-scaling-stroke;
	}
</style>
