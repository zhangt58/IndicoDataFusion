// Local router module - provides simple client-side routing for Wails apps
import Router from './Router.svelte';

export { Router };

export function goto(path) {
  try {
    window.history.pushState({}, '', path);
    window.dispatchEvent(new PopStateEvent('popstate'));
    return Promise.resolve();
  } catch (e) {
    return Promise.reject(e);
  }
}
