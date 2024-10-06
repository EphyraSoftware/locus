export interface ApiError {
  error: string
}

export interface UserName {
  username: string
}

export interface UserDetails {
  email: string
  name: UserName
}

export interface User {
  user: UserDetails
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
  createdAt: number
  deleted: boolean
  deletedAt?: number
}

export interface PasscodeResponse {
  passcode: string
  nextPasscode: string
  serverTime: number
  period: number
}

export interface BackupWarning {
  lastBackupAt?: number
  numberNotBackedUp: number
}
