export interface ApiError {
  error: string
}

export interface CodeGroup {
  group_id: string
  name: string
  codes?: CodeSummary[]
}

export interface CodeSummary {
  code_id: string
  name: string
  preferred_name?: string
}
