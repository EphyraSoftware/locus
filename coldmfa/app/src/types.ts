export interface ApiError {
  error: string
}

export interface CodeGroup {
  groupId: string
  name: string
  codes?: CodeSummary[]
}

export interface CodeSummary {
  codeId: string
  name: string
  preferredName?: string
}
