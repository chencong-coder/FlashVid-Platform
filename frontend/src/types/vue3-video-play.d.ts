declare module 'vue3-video-play' {
  import type { DefineComponent } from 'vue'

  interface VideoPlayProps {
    width?: string
    height?: string
    src: string
    poster?: string
    muted?: boolean
    autoPlay?: boolean
    loop?: boolean
    volume?: number
    control?: boolean
    playsinline?: boolean
  }

  export const videoPlay: DefineComponent<VideoPlayProps>
}
