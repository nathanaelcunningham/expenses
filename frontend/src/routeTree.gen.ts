/* eslint-disable */

// @ts-nocheck

// noinspection JSUnusedGlobalSymbols

// This file was automatically generated by TanStack Router.
// You should NOT make any changes in this file as it will be overwritten.
// Additionally, you should also exclude this file from your linter and/or formatter to prevent it from being checked or modified.

// Import Routes

import { Route as rootRoute } from './routes/__root'
import { Route as IndexImport } from './routes/index'
import { Route as ExpenseCreateImport } from './routes/expense.create'
import { Route as ExpenseIdEditImport } from './routes/expense.$id.edit'

// Create/Update Routes

const IndexRoute = IndexImport.update({
  id: '/',
  path: '/',
  getParentRoute: () => rootRoute,
} as any)

const ExpenseCreateRoute = ExpenseCreateImport.update({
  id: '/expense/create',
  path: '/expense/create',
  getParentRoute: () => rootRoute,
} as any)

const ExpenseIdEditRoute = ExpenseIdEditImport.update({
  id: '/expense/$id/edit',
  path: '/expense/$id/edit',
  getParentRoute: () => rootRoute,
} as any)

// Populate the FileRoutesByPath interface

declare module '@tanstack/react-router' {
  interface FileRoutesByPath {
    '/': {
      id: '/'
      path: '/'
      fullPath: '/'
      preLoaderRoute: typeof IndexImport
      parentRoute: typeof rootRoute
    }
    '/expense/create': {
      id: '/expense/create'
      path: '/expense/create'
      fullPath: '/expense/create'
      preLoaderRoute: typeof ExpenseCreateImport
      parentRoute: typeof rootRoute
    }
    '/expense/$id/edit': {
      id: '/expense/$id/edit'
      path: '/expense/$id/edit'
      fullPath: '/expense/$id/edit'
      preLoaderRoute: typeof ExpenseIdEditImport
      parentRoute: typeof rootRoute
    }
  }
}

// Create and export the route tree

export interface FileRoutesByFullPath {
  '/': typeof IndexRoute
  '/expense/create': typeof ExpenseCreateRoute
  '/expense/$id/edit': typeof ExpenseIdEditRoute
}

export interface FileRoutesByTo {
  '/': typeof IndexRoute
  '/expense/create': typeof ExpenseCreateRoute
  '/expense/$id/edit': typeof ExpenseIdEditRoute
}

export interface FileRoutesById {
  __root__: typeof rootRoute
  '/': typeof IndexRoute
  '/expense/create': typeof ExpenseCreateRoute
  '/expense/$id/edit': typeof ExpenseIdEditRoute
}

export interface FileRouteTypes {
  fileRoutesByFullPath: FileRoutesByFullPath
  fullPaths: '/' | '/expense/create' | '/expense/$id/edit'
  fileRoutesByTo: FileRoutesByTo
  to: '/' | '/expense/create' | '/expense/$id/edit'
  id: '__root__' | '/' | '/expense/create' | '/expense/$id/edit'
  fileRoutesById: FileRoutesById
}

export interface RootRouteChildren {
  IndexRoute: typeof IndexRoute
  ExpenseCreateRoute: typeof ExpenseCreateRoute
  ExpenseIdEditRoute: typeof ExpenseIdEditRoute
}

const rootRouteChildren: RootRouteChildren = {
  IndexRoute: IndexRoute,
  ExpenseCreateRoute: ExpenseCreateRoute,
  ExpenseIdEditRoute: ExpenseIdEditRoute,
}

export const routeTree = rootRoute
  ._addFileChildren(rootRouteChildren)
  ._addFileTypes<FileRouteTypes>()

/* ROUTE_MANIFEST_START
{
  "routes": {
    "__root__": {
      "filePath": "__root.tsx",
      "children": [
        "/",
        "/expense/create",
        "/expense/$id/edit"
      ]
    },
    "/": {
      "filePath": "index.tsx"
    },
    "/expense/create": {
      "filePath": "expense.create.tsx"
    },
    "/expense/$id/edit": {
      "filePath": "expense.$id.edit.tsx"
    }
  }
}
ROUTE_MANIFEST_END */
