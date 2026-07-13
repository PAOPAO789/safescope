import { describe, expect, it } from "vitest";
import { formatDate, labelize, matchesAsset, matchesProject } from "../lib/api";

describe("api presentation helpers", () => {
  it("humanizes enum labels", () => {
    expect(labelize("archived_projects")).toBe("Archived Projects");
  });

  it("formats ISO dates", () => {
    expect(formatDate("2026-01-15T12:00:00Z")).toContain("2026");
  });

  it("searches asset values, status, type, and tags", () => {
    const asset = { value: "api.example.com", type: "domain", status: "alive", tags: ["production"] } as const;
    expect(matchesAsset(asset, "PROD")).toBe(true);
    expect(matchesAsset(asset, "down")).toBe(false);
  });

  it("searches project names, descriptions, and status", () => {
    const project = { name: "External perimeter", description: "Public attack surface", status: "active" } as const;
    expect(matchesProject(project, "attack")).toBe(true);
    expect(matchesProject(project, "archived")).toBe(false);
  });
});
