import adapter from '@sveltejs/adapter-static';
import cspDirectives from './csp-directives.mjs';

export default {
	kit: {
		adapter: adapter({
			fallback: 'index.html' // may differ from host to host
		}),
		csp: {
			mode: "hash",
			directives: cspDirectives
		},
	}
};

