import requests from '../utils/request'

export interface UploadedFile {
  id: number | string
  file: File
  uploadedSize: number
  status: 'pending' | 'uploading' | 'uploaded'
  cancel: AbortController
}

export function uploadDocument<T extends UploadedFile>(
  file: T,
  onChange?: (f: UploadedFile) => void
) {
  const data = new FormData()

  data.append('file', file.file)

  const promise = requests.postForm('/ui/api/documents/upload', data, {
    timeout: 5 * 60 * 1000, // 5mins
    signal: file.cancel.signal,
    onUploadProgress: (event) => {
      if (file.status === 'pending') {
        file.status = 'uploading'
      }
      const { loaded, total } = event

      file.uploadedSize = loaded
      if (loaded === total) {
        file.status = 'uploaded'
      }
      onChange && onChange(file)
    }
  })

  return promise
}
