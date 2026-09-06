/**
 * Helpers puros de la feature `guide`.
 *
 * Los datos del manual vienen de `manual-meta.ts`, que genera el build; aquí
 * solo está la presentación de esos datos.
 */
import type { ManualMeta } from './manual-meta';

/** Un capítulo del manual, tal y como lo escribe `pnpm manual:build`. */
export type ManualSection = ManualMeta['sections'][number];

/** Un bloque del índice: cómo se llama y qué capítulos caen dentro. */
export interface ManualGroup {
	label: string;
	sections: ManualSection[];
}

/*
 * Los cuatro bloques en que se reparte el índice, por el capítulo en que
 * empieza cada uno.
 *
 * Diecinueve títulos seguidos no son un índice, son una pared: para saber si lo
 * tuyo está dentro hay que leerlos todos. Agrupados, la pregunta pasa a ser de
 * cuál de cuatro cosas se trata, y eso se responde de un vistazo.
 *
 * Va por número de capítulo y no por título para que el índice siga funcionando
 * cuando el manual cambie: un capítulo renombrado no se descuelga, y uno nuevo
 * cae en el bloque de los que tiene alrededor. Solo hay que tocar esta lista si
 * el manual se reorganiza de arriba abajo, que es justo cuando toca revisarla.
 */
const GROUPS: { from: number; label: string }[] = [
	{ from: 1, label: 'Para empezar' },
	{ from: 5, label: 'El día a día' },
	{ from: 13, label: 'Tu cuenta' },
	{ from: 17, label: 'Si te atascas' }
];

/**
 * Reparte los capítulos del manual en los bloques del índice.
 *
 * Cada capítulo va al último bloque que empieza en su número o antes, así que
 * los que aparezcan por debajo del primero —o por encima del último— tienen
 * sitio igual. Los bloques que se queden sin capítulos no se pintan.
 */
export function groupSections(sections: ManualSection[]): ManualGroup[] {
	const groups: ManualGroup[] = GROUPS.map(({ label }) => ({ label, sections: [] }));

	for (const section of sections) {
		let index = 0;
		for (let i = 0; i < GROUPS.length; i++) {
			if (section.number >= GROUPS[i].from) index = i;
		}
		groups[index].sections.push(section);
	}

	return groups.filter((group) => group.sections.length > 0);
}

/** Tamaño del PDF en la unidad que toque, para poder decidir si descargarlo. */
export function formatBytes(bytes: number): string {
	if (!Number.isFinite(bytes) || bytes <= 0) return '—';
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
	// Con coma decimal: la cifra va dentro de una frase en español, y «2.6 MB»
	// entre prosa en castellano se lee como dos mil seiscientos.
	const megabytes = (bytes / (1024 * 1024)).toLocaleString('es-CO', {
		minimumFractionDigits: 1,
		maximumFractionDigits: 1
	});
	return `${megabytes} MB`;
}

/** Fecha de generación en formato largo; el ISO crudo no dice nada al usuario. */
export function formatGeneratedAt(iso: string): string {
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) return '—';
	// Día sin cero delante: la fecha va dentro de una frase, y «el 05 de
	// septiembre» no es como se lee una fecha en voz alta.
	return date.toLocaleDateString('es-CO', { day: 'numeric', month: 'long', year: 'numeric' });
}
