// themeStore.js
import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';

function createThemeStore() {
  const initialValue = browser ? localStorage.getItem('theme') || 'dark' : 'light';
  console.log("theme color")
  console.log(initialValue)
  if (browser) {
    document.body.setAttribute('data-bs-theme', initialValue);
  }
  const { subscribe, set } = writable(initialValue);

  return {
    subscribe,
    set: (value) => {
      if (browser) {
        localStorage.setItem('theme', value);
        const event = new CustomEvent('themeChange', { detail: value });
        window.dispatchEvent(event);
        document.body.setAttribute('data-bs-theme', event.detail);
      }
      set(value);
    }
  };
}

export const theme = createThemeStore();
