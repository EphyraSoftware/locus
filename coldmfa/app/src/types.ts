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
