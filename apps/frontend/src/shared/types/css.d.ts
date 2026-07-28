// Metro's isCSSEnabled (see metro.config.js) lets web builds import .css
// directly; TypeScript needs an ambient module declaration to allow it.
declare module '*.css'
