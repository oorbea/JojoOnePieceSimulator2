// Centralizes static asset requires so feature screens don't need brittle
// multi-level relative paths (the old login-screen.tsx reached up 5
// directories to get here).
export const logoAsset = require('../../assets/images/logo.png')
