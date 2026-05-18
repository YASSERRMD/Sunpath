/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_TILE_STYLE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
