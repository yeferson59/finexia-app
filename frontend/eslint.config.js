import prettier from 'eslint-config-prettier';
import path from 'node:path';
import { includeIgnoreFile } from 'eslint/config';
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import { defineConfig } from 'eslint/config';
import globals from 'globals';
import ts from 'typescript-eslint';
import svelteConfig from './svelte.config.js';

const gitignorePath = path.resolve(import.meta.dirname, '.gitignore');

export default defineConfig(
	includeIgnoreFile(gitignorePath),
	js.configs.recommended,
	ts.configs.recommended,
	svelte.configs.recommended,
	prettier,
	svelte.configs.prettier,
	{
		languageOptions: { globals: { ...globals.browser, ...globals.node } },
		rules: {
			// typescript-eslint strongly recommend that you do not use the no-undef lint rule on TypeScript projects.
			// see: https://typescript-eslint.io/troubleshooting/faqs/eslint/#i-get-errors-from-the-no-undef-rule-about-global-variables-not-being-defined-even-though-there-are-no-typescript-errors
			'no-undef': 'off',
			// Allow intentionally unused bindings prefixed with `_` (e.g. `{#each items as _, i}`).
			'@typescript-eslint/no-unused-vars': [
				'error',
				{
					argsIgnorePattern: '^_',
					varsIgnorePattern: '^_',
					caughtErrorsIgnorePattern: '^_'
				}
			]
		}
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
		languageOptions: {
			parserOptions: {
				projectService: true,
				extraFileExtensions: ['.svelte'],
				parser: ts.parser,
				svelteConfig
			}
		}
	},
	// ---------------------------------------------------------------------
	// Fronteras de la arquitectura (docs/FRONTEND_ARCHITECTURE.md).
	//
	// Las capas van de abajo arriba: shared → ui → api → features → routes.
	// Cada bloque prohíbe los imports que irían en sentido contrario o que
	// saltarían la superficie pública de una feature. Hasta ahora eran una
	// convención escrita; aquí fallan el CI.
	// ---------------------------------------------------------------------
	{
		files: ['src/lib/shared/**/*.{ts,js,svelte}'],
		rules: {
			'no-restricted-imports': [
				'error',
				{
					patterns: [
						{
							group: ['$lib/features', '$lib/features/**', '$lib/api', '$lib/api/**', '$lib/ui/**'],
							message:
								'lib/shared es la capa más baja: no importa de features, api ni ui (docs/FRONTEND_ARCHITECTURE.md).'
						}
					]
				}
			]
		}
	},
	{
		files: ['src/lib/ui/**/*.{ts,js,svelte}'],
		rules: {
			'no-restricted-imports': [
				'error',
				{
					patterns: [
						{
							group: [
								'$lib/features',
								'$lib/features/**',
								'$lib/api',
								'$lib/api/**',
								'../features/**',
								'../api/**'
							],
							message:
								'lib/ui es el design system: no conoce dominios. Recibe los datos por props o snippets.'
						}
					]
				}
			]
		}
	},
	{
		files: ['src/lib/api/**/*.{ts,js,svelte}'],
		rules: {
			'no-restricted-imports': [
				'error',
				{
					patterns: [
						{
							group: ['$lib/features', '$lib/features/**', '$lib/ui/**', '../features/**'],
							message:
								'lib/api es la capa de acceso al backend: no depende de la UI ni de features.'
						}
					]
				}
			]
		}
	},
	{
		files: ['src/lib/features/**/*.{ts,js,svelte}'],
		rules: {
			'no-restricted-imports': [
				'error',
				{
					patterns: [
						{
							// Dentro de una feature todo se importa en relativo, así que un
							// `$lib/features/...` (o un `../../otra`) solo puede ser otra feature.
							group: ['$lib/features/**', '../../*', '../../*/**'],
							message:
								'Una feature no importa de otra feature: lo compartido baja a lib/ui, lib/api o lib/shared.'
						}
					]
				}
			]
		}
	},
	{
		files: ['src/routes/**/*.{ts,js,svelte}'],
		rules: {
			'no-restricted-imports': [
				'error',
				{
					patterns: [
						{
							// La hoja de estilos de una feature es la excepción documentada:
							// un barrel de JS no puede reexportar un side-effect de CSS.
							group: ['$lib/features/*/**', '!$lib/features/*/*.css'],
							message:
								'routes/ importa la superficie pública de la feature ($lib/features/<x>), no sus internos.'
						},
						{
							group: ['$lib/api/client'],
							message:
								'routes/ no llama a authedFetch directamente: usa un módulo de dominio de lib/api.'
						}
					]
				}
			]
		}
	}
);
