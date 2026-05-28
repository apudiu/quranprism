import {
  createSolidTable,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  type ColumnDef,
  type SortingState,
  type Updater,
} from "@tanstack/solid-table";
import ArrowDown from "lucide-solid/icons/arrow-down";
import ArrowUp from "lucide-solid/icons/arrow-up";
import ChevronsUpDown from "lucide-solid/icons/chevrons-up-down";
import { createSignal, For, Show, type JSX } from "solid-js";
import { cn } from "../../lib/utils";
import { Button } from "../ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";

export interface DataTableProps<TData, TValue> {
  data: TData[];
  columns: ColumnDef<TData, TValue>[];
  loading?: boolean;
  pageSize?: number;
  searchable?: boolean;
  searchPlaceholder?: string;
  /** Controlled search value (e.g. URL-synced). Omit for internal state. */
  search?: string;
  onSearchChange?: (value: string) => void;
  /** Extra toolbar content (e.g. an "Add" button), right-aligned. */
  toolbar?: JSX.Element;
  emptyMessage?: string;
}

const resolve = <T,>(updater: Updater<T>, prev: T): T =>
  typeof updater === "function" ? (updater as (p: T) => T)(prev) : updater;

/** Generic client-side table (sort + filter + paginate) on TanStack Solid Table
 * with solid-ui styling. Server-mode + URL sync layer on top via useDataTable. */
export function DataTable<TData, TValue>(props: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = createSignal<SortingState>([]);
  const [internalFilter, setInternalFilter] = createSignal("");

  const filter = () => props.search ?? internalFilter();
  const setFilter = (v: string) =>
    props.onSearchChange ? props.onSearchChange(v) : setInternalFilter(v);

  const table = createSolidTable({
    get data() {
      return props.data;
    },
    get columns() {
      return props.columns;
    },
    state: {
      get sorting() {
        return sorting();
      },
      get globalFilter() {
        return filter();
      },
    },
    onSortingChange: (u) => setSorting((p) => resolve(u, p)),
    onGlobalFilterChange: (u) => setFilter(resolve(u, filter())),
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: { pagination: { pageSize: props.pageSize ?? 10 } },
  });

  return (
    <div class="space-y-4">
      <Show when={props.searchable || props.toolbar}>
        <div class="flex items-center justify-between gap-2">
          <Show when={props.searchable}>
            <input
              type="search"
              value={filter()}
              onInput={(e) => setFilter(e.currentTarget.value)}
              placeholder={props.searchPlaceholder ?? "Search…"}
              class="h-9 w-64 max-w-full rounded-md border border-input bg-transparent px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </Show>
          <div class="ml-auto flex items-center gap-2">{props.toolbar}</div>
        </div>
      </Show>

      <div class="rounded-lg border border-border">
        <Table>
          <TableHeader>
            <For each={table.getHeaderGroups()}>
              {(hg) => (
                <TableRow>
                  <For each={hg.headers}>
                    {(header) => {
                      const sortable = header.column.getCanSort();
                      return (
                        <TableHead>
                          <Show when={!header.isPlaceholder}>
                            <button
                              type="button"
                              disabled={!sortable}
                              onClick={header.column.getToggleSortingHandler()}
                              class={cn(
                                "flex items-center gap-1",
                                sortable && "cursor-pointer select-none hover:text-foreground",
                              )}
                            >
                              {flexRender(header.column.columnDef.header, header.getContext())}
                              <Show when={sortable}>
                                {header.column.getIsSorted() === "asc" ? (
                                  <ArrowUp class="size-3.5" />
                                ) : header.column.getIsSorted() === "desc" ? (
                                  <ArrowDown class="size-3.5" />
                                ) : (
                                  <ChevronsUpDown class="size-3.5 opacity-50" />
                                )}
                              </Show>
                            </button>
                          </Show>
                        </TableHead>
                      );
                    }}
                  </For>
                </TableRow>
              )}
            </For>
          </TableHeader>
          <TableBody>
            <Show
              when={!props.loading}
              fallback={
                <TableRow>
                  <TableCell colSpan={props.columns.length} class="h-24 text-center text-muted-foreground">
                    Loading…
                  </TableCell>
                </TableRow>
              }
            >
              <Show
                when={table.getRowModel().rows.length}
                fallback={
                  <TableRow>
                    <TableCell colSpan={props.columns.length} class="h-24 text-center text-muted-foreground">
                      {props.emptyMessage ?? "No results."}
                    </TableCell>
                  </TableRow>
                }
              >
                <For each={table.getRowModel().rows}>
                  {(row) => (
                    <TableRow>
                      <For each={row.getVisibleCells()}>
                        {(cell) => (
                          <TableCell>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                        )}
                      </For>
                    </TableRow>
                  )}
                </For>
              </Show>
            </Show>
          </TableBody>
        </Table>
      </div>

      <div class="flex items-center justify-between">
        <p class="text-sm text-muted-foreground">
          {table.getFilteredRowModel().rows.length} row(s)
        </p>
        <div class="flex items-center gap-2">
          <span class="text-sm text-muted-foreground">
            Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount() || 1}
          </span>
          <Button variant="outline" size="sm" disabled={!table.getCanPreviousPage()} onClick={() => table.previousPage()}>
            Previous
          </Button>
          <Button variant="outline" size="sm" disabled={!table.getCanNextPage()} onClick={() => table.nextPage()}>
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
