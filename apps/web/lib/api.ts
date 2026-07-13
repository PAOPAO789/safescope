import type { Asset, Project } from "./types";

export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public code = "request_failed",
  ) {
    super(message);
  }
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/backend${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
    cache: "no-store",
  });

  if (response.status === 401 && typeof window !== "undefined") {
    window.location.assign("/login");
    throw new APIError("Session expired", 401, "unauthorized");
  }
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new APIError(
      body?.error?.message ?? "Request failed",
      response.status,
      body?.error?.code,
    );
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return response.json() as Promise<T>;
}

export function formatDate(value: string): string {
  return new Intl.DateTimeFormat("en", {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}

export function labelize(value: string): string {
  return value.replaceAll("_", " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}

export function matchesAsset(asset: Pick<Asset, "value" | "type" | "status" | "tags">, query: string): boolean {
  const term = query.trim().toLowerCase();
  return !term || [asset.value, asset.type, asset.status, ...asset.tags].some((value) => value.toLowerCase().includes(term));
}

export function matchesProject(project: Pick<Project, "name" | "description" | "status">, query: string): boolean {
  const term = query.trim().toLowerCase();
  return !term || [project.name, project.description, project.status].some((value) => value.toLowerCase().includes(term));
}
