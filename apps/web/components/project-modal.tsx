"use client";

import { FormEvent, useState } from "react";
import { X } from "lucide-react";
import { api } from "@/lib/api";
import type { Project } from "@/lib/types";

export function ProjectModal({ onClose, onCreated }: { onClose: () => void; onCreated: (project: Project) => void }) {
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    try {
      const form = Object.fromEntries(new FormData(event.currentTarget));
      const project = await api<Project>("/projects", { method: "POST", body: JSON.stringify(form) });
      onCreated(project);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to create project");
      setLoading(false);
    }
  }

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <div className="modal" role="dialog" aria-modal="true" aria-labelledby="project-modal-title">
        <div className="modal-head"><h2 id="project-modal-title">New project</h2><button className="icon-plain" onClick={onClose} aria-label="Close"><X size={19} /></button></div>
        <form className="modal-body" onSubmit={submit}>
          <div className="field"><label htmlFor="project-name">Name</label><input id="project-name" name="name" required autoFocus /></div>
          <div className="field"><label htmlFor="project-description">Description</label><textarea id="project-description" name="description" /></div>
          {error && <p className="form-error">{error}</p>}
          <div className="modal-actions"><button className="button secondary" type="button" onClick={onClose}>Cancel</button><button className="button" disabled={loading}>Create project</button></div>
        </form>
      </div>
    </div>
  );
}
