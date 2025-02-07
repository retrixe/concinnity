// https://web.dev/patterns/files/open-one-or-multiple-files#progressive_enhancement
interface FileWithHandle extends File {
  handle?: FileSystemFileHandle
}

export async function openFileOrFiles<T extends boolean = false>(options?: {
  multiple?: T
  types?: { description?: string; accept: Record<string, string[]> }[]
}): Promise<(T extends true ? FileWithHandle[] : FileWithHandle) | undefined> {
  // Feature detection. The API needs to be supported
  // and the app not run in an iframe.
  const supportsFileSystemAccess =
    'showOpenFilePicker' in window &&
    (() => {
      try {
        return window.self === window.top
      } catch {
        return false
      }
    })()
  // If the File System Access API is supported…
  if (supportsFileSystemAccess) {
    let fileOrFiles: FileWithHandle | FileWithHandle[] | undefined = undefined
    try {
      // Show the file picker, optionally allowing multiple files.
      // @ts-expect-error -- Seems to not be recognized by TypeScript.
      // eslint-disable-next-line
      const handles: FileSystemFileHandle[] = await showOpenFilePicker(options)
      // Only one file is requested.
      if (!options?.multiple) {
        // Add the `FileSystemFileHandle` as `.handle`.
        fileOrFiles = await handles[0].getFile()
        fileOrFiles.handle = handles[0]
      } else {
        fileOrFiles = await Promise.all(
          handles.map(async handle => {
            const file: FileWithHandle = await handle.getFile()
            // Add the `FileSystemFileHandle` as `.handle`.
            file.handle = handle
            return file
          }),
        )
      }
    } catch (err: unknown) {
      // Fail silently if the user has simply canceled the dialog.
      if (!(err instanceof Error) || err.name !== 'AbortError') {
        throw err
      }
    }
    return fileOrFiles as (T extends true ? FileWithHandle[] : FileWithHandle) | undefined
  }
  // Fallback if the File System Access API is not supported.
  return new Promise(resolve => {
    // Append a new `<input type="file" multiple? />` and hide it.
    const input = document.createElement('input')
    input.style.display = 'none'
    input.type = 'file'
    document.body.append(input)
    if (options?.multiple) {
      input.multiple = true
    }
    if (options?.types?.length) {
      input.accept = options.types.flatMap(({ accept }) => Object.entries(accept)).join(',')
    }
    // The `change` event fires when the user interacts with the dialog.
    input.addEventListener('change', () => {
      // Remove the `<input type="file" multiple? />` again from the DOM.
      input.remove()
      // If no files were selected, return.
      if (!input.files) {
        return
      }
      // Return all files or just one file.
      const fileOrFiles = options?.multiple ? Array.from(input.files) : input.files[0]
      resolve(fileOrFiles as (T extends true ? FileWithHandle[] : FileWithHandle) | undefined)
    })
    // Show the picker.
    if ('showPicker' in HTMLInputElement.prototype) {
      input.showPicker()
    } else {
      input.click()
    }
  })
}
