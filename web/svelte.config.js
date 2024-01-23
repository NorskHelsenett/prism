import adapter from '@sveltejs/adapter-static';
import cspDirectives from './csp-directives.mjs';
const production = process.env.NODE_ENV === 'production';

export default {
	kit: {
		adapter: adapter({
			fallback: 'index.html' // may differ from host to host
		}),
		...(production ? {
			// Production only configuration
			csp: {
				mode: "hash",
				directives: cspDirectives
			},
		} : {
			// Development only configuration
		})
	}
};

