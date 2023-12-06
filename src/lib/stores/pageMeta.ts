// src/stores/pageMeta.js
import { writable } from 'svelte/store';

export const pageMeta = writable({
  title: 'PRISM',
  pretitle: 'Pentest Report Information Security Management'
});
