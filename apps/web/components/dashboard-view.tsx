"use client";

import { useEffect, useState } from "react";
import { Activity, Archive, Boxes, RadioTower } from "lucide-react";
import { api, labelize } from "@/lib/api";
import type { Dashboard } from "@/lib/types";

const initial: Dashboard = {
  projects: 0, assets: 0, alive_assets: 0, archived_projects: 0, recent_assets: [], assets_by_type: [],
};

export function DashboardView() {
  const [data, setData] = useState<Dashboard>(initial);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api<Dashboard>("/dashboard").then(setData).finally(() => setLoading(false));
  }, []);

  const max = Math.max(...data.assets_by_type.map((metric) => metric.value), 1);
  const metrics = [
    { label: "Active projects", value: data.projects - data.archived_projects, icon: Boxes, note: `${data.archived_projects} archived` },
    { label: "Managed assets", value: data.assets, icon: RadioTower, note: "Across all scopes" },
    { label: "Responsive assets", value: data.alive_assets, icon: Activity, note: data.assets ? `${Math.round((data.alive_assets / data.assets) * 100)}% observed alive` : "No observations yet" },
    { label: "Archived projects", value: data.archived_projects, icon: Archive, note: "Retained for history" },
  ];

  return (
    <div className="content">
      <div className="page-head">
        <div><h1>Security overview</h1><p>Current asset visibility and workspace activity.</p></div>
      </div>
      <section className="metrics">
        {metrics.map(({ label, value, icon: Icon, note }) => (
          <div className="metric-card" key={label}>
            <div className="metric-label"><span>{label}</span><Icon size={16} /></div>
            <strong>{loading ? "..." : value}</strong><small>{note}</small>
          </div>
        ))}
      </section>
      <div className="dashboard-grid">
        <section className="panel">
          <div className="panel-head"><h2>Recently updated assets</h2></div>
          {data.recent_assets.length ? (
            <div className="table-wrap">
              <table>
                <thead><tr><th>Asset</th><th>Type</th><th>Status</th><th>Tags</th></tr></thead>
                <tbody>{data.recent_assets.map((asset) => (
                  <tr key={asset.id}>
                    <td className="asset-value">{asset.value}</td>
                    <td>{labelize(asset.type)}</td>
                    <td><span className={`badge ${asset.status}`}><span className="status-dot" />{asset.status}</span></td>
                    <td><div className="tags">{asset.tags.slice(0, 2).map((tag) => <span className="tag" key={tag}>{tag}</span>)}</div></td>
                  </tr>
                ))}</tbody>
              </table>
            </div>
          ) : <div className="empty">No assets have been added yet.</div>}
        </section>
        <section className="panel">
          <div className="panel-head"><h2>Asset distribution</h2></div>
          <div className="panel-body distribution">
            {data.assets_by_type.length ? data.assets_by_type.map((metric) => (
              <div className="bar-row" key={metric.label}>
                <div className="bar-label"><span>{labelize(metric.label)}</span><strong>{metric.value}</strong></div>
                <div className="bar-track"><div className="bar-fill" style={{ width: `${(metric.value / max) * 100}%` }} /></div>
              </div>
            )) : <div className="empty">Distribution appears after assets are added.</div>}
          </div>
        </section>
      </div>
    </div>
  );
}
