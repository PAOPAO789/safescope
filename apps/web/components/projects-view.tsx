"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { api, formatDate, matchesProject } from "@/lib/api";
import type { Project, User } from "@/lib/types";
import { ProjectModal } from "./project-modal";

export function ProjectsView() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");

  useEffect(() => {
    Promise.all([api<{ items: Project[] }>("/projects"), api<User>("/me")])
      .then(([result, current]) => { setProjects(result.items); setUser(current); })
      .finally(() => setLoading(false));
  }, []);

  const visibleProjects = projects.filter((project) => matchesProject(project, query));

  return (
    <div className="content">
      <div className="page-head">
        <div><h1>Projects</h1><p>Organize assets into clear engagement and ownership scopes.</p></div>
        <div className="page-actions">
          <input className="panel-search" type="search" aria-label="Search projects" placeholder="Search projects" value={query} onChange={(event) => setQuery(event.target.value)} />
          {user?.role !== "viewer" && <button className="button" onClick={() => setOpen(true)}><Plus size={16} />New project</button>}
        </div>
      </div>
      {loading ? <div className="loading">Loading projects...</div> : visibleProjects.length ? (
        <section className="project-grid">{visibleProjects.map((project) => (
          <Link className="project-card" href={`/projects/${project.id}`} key={project.id}>
            <div className="project-card-head"><h2>{project.name}</h2><span className={`badge ${project.status}`}>{project.status}</span></div>
            <p>{project.description || "No project description."}</p>
            <div className="project-stats"><span><strong>{project.asset_count}</strong> assets</span><span>Updated {formatDate(project.updated_at)}</span></div>
          </Link>
        ))}</section>
      ) : <div className="panel empty">{projects.length ? "No projects match this search." : "Create your first project to start tracking assets."}</div>}
      {open && <ProjectModal onClose={() => setOpen(false)} onCreated={(project) => { setProjects((items) => [project, ...items]); setOpen(false); }} />}
    </div>
  );
}
