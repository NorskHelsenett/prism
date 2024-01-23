import adapter from '@sveltejs/adapter-static';

export default {
	kit: {
		adapter: adapter({
			fallback: 'index.html' // may differ from host to host
		}),
		csp: {
			mode: "hash",
			directives: {
				"script-src": ["self"],
			}
		},
	}
};

