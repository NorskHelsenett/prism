import { writable } from 'svelte/store';
import { browser } from '$app/environment';

function createThemeStore() {
  const initialValue = browser ? localStorage.getItem('theme') || 'light' : 'light';
  const { subscribe, set } = writable(initialValue);

  return {
    subscribe,
    set: (value) => {
      if (browser) {
        localStorage.setItem('theme', value);
        const event = new CustomEvent('themeChange', { detail: value });
        window.dispatchEvent(event);
      }
      set(value);
    }
  };
}

export const theme = createThemeStore();
