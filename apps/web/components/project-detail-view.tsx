"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, Pencil, Plus, Radar, Settings, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { api, formatDate, labelize, matchesAsset } from "@/lib/api";
import type { Asset, Project, User } from "@/lib/types";
import { AssetModal } from "./asset-modal";
import { ProjectSettingsModal } from "./project-settings-modal";

export function ProjectDetailView({ id }: { id: string }) {
  const router = useRouter();
  const [project, setProject] = useState<Project | null>(null);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [editing, setEditing] = useState<Asset | "new" | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [query, setQuery] = useState("");

  useEffect(() => {
    Promise.all([
      api<Project>(`/projects/${id}`),
      api<{ items: Asset[] }>(`/projects/${id}/assets`),
      api<User>("/me"),
    ]).then(([projectResult, assetResult, current]) => {
      setProject(projectResult);
      setAssets(assetResult.items);
      setUser(current);
    });
  }, [id]);

  async function remove(asset: Asset) {
    if (!window.confirm(`Delete ${asset.value}?`)) return;
    await api<void>(`/assets/${asset.id}`, { method: "DELETE" });
    setAssets((items) => items.filter((item) => item.id !== asset.id));
  }

  if (!project) return <div className="content loading">Loading project...</div>;
  const writable = user?.role !== "viewer";
  const visibleAssets = assets.filter((asset) => matchesAsset(asset, query));
  return (
    <div className="content">
      <div className="page-head">
        <div>
          <Link href="/projects" className="nav-link" style={{ padding: 0, minHeight: 28, color: "var(--muted)" }}><ArrowLeft size={15} />Projects</Link>
          <h1>{project.name}</h1>
          <p>{project.description || "No project description."}</p>
        </div>
        <div className="page-actions">
          <button className="button secondary" disabled title="Scanner adapters are ready for future configuration"><Radar size={16} />Run scan</button>
          {writable && <button className="button secondary" onClick={() => setSettingsOpen(true)} title="Project settings"><Settings size={16} />Settings</button>}
          {writable && <button className="button" onClick={() => setEditing("new")}><Plus size={16} />Add asset</button>}
        </div>
      </div>
      <section className="metrics">
        <div className="metric-card"><div className="metric-label">Total assets</div><strong>{assets.length}</strong><small>In this project</small></div>
        <div className="metric-card"><div className="metric-label">Alive</div><strong>{assets.filter((asset) => asset.status === "alive").length}</strong><small>Responsive targets</small></div>
        <div className="metric-card"><div className="metric-label">Domains</div><strong>{assets.filter((asset) => asset.type === "domain").length}</strong><small>DNS scopes</small></div>
        <div className="metric-card"><div className="metric-label">Updated</div><strong style={{ fontSize: 18 }}>{formatDate(project.updated_at)}</strong><small>Project activity</small></div>
      </section>
      <section className="panel">
        <div className="panel-head">
          <h2>Assets</h2>
          <div className="panel-tools">
            <input className="panel-search" type="search" aria-label="Search assets" placeholder="Search assets" value={query} onChange={(event) => setQuery(event.target.value)} />
            <span className={`badge ${project.status}`}>{project.status}</span>
          </div>
        </div>
        {visibleAssets.length ? (
          <div className="table-wrap"><table>
            <thead><tr><th>Value</th><th>Type</th><th>Status</th><th>Tags</th><th>Updated</th>{writable && <th>Actions</th>}</tr></thead>
            <tbody>{visibleAssets.map((asset) => (
              <tr key={asset.id}>
                <td className="asset-value">{asset.value}</td>
                <td>{labelize(asset.type)}</td>
                <td><span className={`badge ${asset.status}`}><span className="status-dot" />{asset.status}</span></td>
                <td><div className="tags">{asset.tags.slice(0, 3).map((tag) => <span className="tag" key={tag}>{tag}</span>)}</div></td>
                <td>{formatDate(asset.updated_at)}</td>
                {writable && <td><div className="page-actions">
                  <button className="icon-plain" title="Edit asset" aria-label="Edit asset" onClick={() => setEditing(asset)}><Pencil size={16} /></button>
                  <button className="icon-plain" title="Delete asset" aria-label="Delete asset" onClick={() => remove(asset)}><Trash2 size={16} /></button>
                </div></td>}
              </tr>
            ))}</tbody>
          </table></div>
        ) : <div className="empty">{assets.length ? "No assets match this search." : "No assets in this scope. Add a domain, IP, URL, or service."}</div>}
      </section>
      {editing && <AssetModal
        projectId={id}
        asset={editing === "new" ? undefined : editing}
        onClose={() => setEditing(null)}
        onSaved={(saved) => {
          setAssets((items) => editing === "new" ? [saved, ...items] : items.map((item) => item.id === saved.id ? saved : item));
          setEditing(null);
        }}
      />}
      {settingsOpen && <ProjectSettingsModal
        project={project}
        onClose={() => setSettingsOpen(false)}
        onSaved={(saved) => { setProject(saved); setSettingsOpen(false); }}
        onDeleted={() => { router.push("/projects"); router.refresh(); }}
      />}
    </div>
  );
}
