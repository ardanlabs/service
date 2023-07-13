// The default service API response.
export interface DefaultAPIResponse<T> {
  page: number
  rowsPerPage: number
  total: number
  items: T[]
}
