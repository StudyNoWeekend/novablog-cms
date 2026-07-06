declare module '@bytemd/vue-next' {
  import type { DefineComponent } from 'vue'
  export const Editor: DefineComponent<any, any, any>
  export const Viewer: DefineComponent<any, any, any>
}

declare module 'bytemd/locales/zh_Hans.json' {
  const value: Record<string, any>
  export default value
}