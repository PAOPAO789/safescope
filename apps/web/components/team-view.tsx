"use client";

import { useEffect, useState } from "react";
import { Shield, Users } from "lucide-react";
import { api, formatDate } from "@/lib/api";
import type { Role, User } from "@/lib/types";

type TeamUser = User & { created_at: string };

export function TeamView() {
  const [users, setUsers] = useState<TeamUser[]>([]);
  const [current, setCurrent] = useState<User | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([api<{ items: TeamUser[] }>("/users"), api<User>("/me")])
      .then(([result, actor]) => { setUsers(result.items); setCurrent(actor); })
      .catch((cause) => setError(cause instanceof Error ? cause.message : "Unable to load team"));
  }, []);

  async function changeRole(user: TeamUser, role: Role) {
    try {
      await api<void>(`/users/${user.id}/role`, { method: "PUT", body: JSON.stringify({ role }) });
      setUsers((items) => items.map((item) => item.id === user.id ? { ...item, role } : item));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to change role");
    }
  }

  return (
    <div className="content">
      <div className="page-head"><div><h1>Team access</h1><p>Manage workspace permissions for operators and observers.</p></div></div>
      <section className="metrics">
        <div className="metric-card"><div className="metric-label">Team members <Users size={16} /></div><strong>{users.length}</strong><small>Registered accounts</small></div>
        <div className="metric-card"><div className="metric-label">Administrators <Shield size={16} /></div><strong>{users.filter((user) => user.role === "admin").length}</strong><small>Full workspace access</small></div>
      </section>
      {error && <p className="form-error">{error}</p>}
      <section className="panel">
        <div className="panel-head"><h2>Workspace members</h2></div>
        <div className="table-wrap"><table>
          <thead><tr><th>Name</th><th>Email</th><th>Joined</th><th>Role</th></tr></thead>
          <tbody>{users.map((user) => (
            <tr key={user.id}>
              <td>{user.name}</td><td>{user.email}</td><td>{formatDate(user.created_at)}</td>
              <td>{user.id === current?.id ? <span className={`badge ${user.role}`}>{user.role} (you)</span> : (
                <select className="role-select" value={user.role} onChange={(event) => changeRole(user, event.target.value as Role)} aria-label={`Role for ${user.name}`}>
                  <option value="admin">Admin</option><option value="analyst">Analyst</option><option value="viewer">Viewer</option>
                </select>
              )}</td>
            </tr>
          ))}</tbody>
        </table></div>
      </section>
    </div>
  );
}
