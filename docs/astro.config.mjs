// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://nguyentin05.github.io',
	base: '/cakd-platform',
	integrations: [
		starlight({
			title: 'CAKD Platform',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/nguyentin05/cakd-platform' }],
			sidebar: [
				{
					label: 'Tutorials',
					items: [{ autogenerate: { directory: 'tutorials' } }],
				},
				{
					label: 'How-to Guides',
					items: [{ autogenerate: { directory: 'how-to-guides' } }],
				},
				{
					label: 'Explanation',
					items: [{ autogenerate: { directory: 'explanation' } }],
				},
				{
					label: 'Reference',
					items: [{ autogenerate: { directory: 'reference' } }],
				},
				{
					label: 'Architecture Decisions (ADRs)',
					items: [{ autogenerate: { directory: 'adrs' } }],
				},
			],
		}),
	],
});
