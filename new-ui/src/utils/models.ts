export interface User {
  userid: string
  name: string
  email?: string
  CreatedAt?: string
  integrations?: string[]
}

type HashDocMode = 'display' | 'editing' | 'creating'

type HashDocType = 'DocumentType' | 'CollectionType'

export interface HashDoc {
  id: string
  name: string
  type: HashDocType
  size: number
  extension?: string
  parent?: string
  children?: HashDoc[]
  lastModified: string

  preMode?: HashDocMode
  mode?: HashDocMode
}

export interface HashDocMetadata {
  visibleName: string
  type: HashDocType
  parent: string
  lastModified: string
  lastOpened: string
  version: number
  pinned: boolean
  synced: boolean
  modified: boolean
  deleted: boolean
  metadatamodified: boolean
}

export interface AppUser {
  userid: string
  email?: string
  name?: string
  CreatedAt?: string
}
