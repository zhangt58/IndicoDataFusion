// Declarations for static asset imports used by the frontend (Svelte + Vite + TypeScript)
// Importing an image file returns the resolved URL as a string.

declare module '*.png' {
  const src: string;
  export default src;
}

declare module '*.jpg' {
  const src: string;
  export default src;
}

declare module '*.jpeg' {
  const src: string;
  export default src;
}

declare module '*.gif' {
  const src: string;
  export default src;
}

declare module '*.webp' {
  const src: string;
  export default src;
}

declare module '*.ico' {
  const src: string;
  export default src;
}

// For SVGs we typically treat them as URLs (string). If you import SVGs as Svelte components,
// update this declaration accordingly (e.g. `export { SvelteComponent }` or use `?component` import plugin).
declare module '*.svg' {
  const src: string;
  export default src;
}

