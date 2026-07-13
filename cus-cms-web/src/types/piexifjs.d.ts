declare module 'piexifjs' {
  export interface ExifValue {
    [tag: number]: number[] | string | number
  }

  export interface ExifObject {
    '0th'?: ExifValue
    '1st'?: ExifValue
    Exif?: ExifValue
    GPS?: ExifValue
    Interop?: ExifValue
    thumbnail?: string
  }

  export const ImageIFD: Record<string, number>
  export const ExifIFD: Record<string, number>
  export const GPSIFD: Record<string, number>

  export function load(jpegBinary: string): ExifObject
  export function dump(exifObj: ExifObject): string
  export function insert(exifDump: string, jpegBinary: string): string
  export function remove(jpegBinary: string): string
}
