"use client";

import { FormEvent, useState } from "react";
import { X } from "lucide-react";
import { api } from "@/lib/api";
import type { Project } from "@/lib/types";

export function ProjectSettingsModal({ project, onClose, onSaved, onDeleted }: {
  project: Project;
  onClose: () => void;
  onSaved: (project: Project) => void;
  onDeleted: () => void;
}) {
  const [error, setError] = useState("");

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      const form = Object.fromEntries(new FormData(event.currentTarget));
      onSaved(await api<Project>(`/projects/${project.id}`, { method: "PUT", body: JSON.stringify(form) }));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to update project");
    }
  }

  async function remove() {
    if (!window.confirm(`Delete ${project.name} and all of its assets?`)) return;
    try {
      await api<void>(`/projects/${project.id}`, { method: "DELETE" });
      onDeleted();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to delete project");
    }
  }

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <div className="modal" role="dialog" aria-modal="true">
        <div className="modal-head"><h2>Project settings</h2><button className="icon-plain" onClick={onClose} aria-label="Close"><X size={19} /></button></div>
        <form className="modal-body" onSubmit={save}>
          <div className="field"><label htmlFor="settings-name">Name</label><input id="settings-name" name="name" defaultValue={project.name} required /></div>
          <div className="field"><label htmlFor="settings-description">Description</label><textarea id="settings-description" name="description" defaultValue={project.description} /></div>
          <div className="field"><label htmlFor="settings-status">Status</label><select id="settings-status" name="status" defaultValue={project.status}><option value="active">Active</option><option value="archived">Archived</option></select></div>
          {error && <p className="form-error">{error}</p>}
          <div className="modal-actions" style={{ justifyContent: "space-between" }}>
            <button className="button danger" type="button" onClick={remove}>Delete project</button>
            <div className="page-actions"><button className="button secondary" type="button" onClick={onClose}>Cancel</button><button className="button">Save changes</button></div>
          </div>
        </form>
      </div>
    </div>
  );
}
