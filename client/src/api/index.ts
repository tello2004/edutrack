// Re-export all API modules

// Client utilities (excluding ApiError which is also in types.ts)
export {
  apiClient,
  getToken,
  setToken,
  removeToken,
  isAuthenticated,
  handleApiError,
} from "./client";

// Types
export * from "./types";

// Auth
export * from "./auth";

// Resources
export * from "./accounts";
export * from "./students";
export * from "./teachers";
export * from "./careers";
export * from "./subjects";
export * from "./attendances";
export * from "./grades";
