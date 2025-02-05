import { writable } from 'svelte/store';

// Check if we're in a browser environment
const prefersDark = typeof window !== 'undefined' 
  ? window.matchMedia('(prefers-color-scheme: dark)').matches 
  : false;

export const theme = writable<'dark' | 'light'>(prefersDark ? 'dark' : 'light'); 

export type Theme = 'dark' | 'light'; 