export interface TableHeader {
  key: string;
  label: string;
  sortable?: boolean;
  format?: (value: any) => string;
}

export interface TableProps {
  data: any[];
  headers: TableHeader[];
  pageSize?: number;
  currentPage?: number;
  totalItems?: number;
  loading?: boolean;
}

export interface SortEvent {
  column: string;
  direction: "asc" | "desc";
}

export interface PageEvent {
  page: number;
}

export interface FilterEvent {
  column: string;
  value: string;
} 