"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Boxes, LayoutDashboard, LogOut, Users } from "lucide-react";
import { api } from "@/lib/api";
import type { User } from "@/lib/types";

const links = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/projects", label: "Projects", icon: Boxes },
];

export function WorkspaceShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    api<User>("/me").then(setUser).catch(() => undefined);
  }, []);

  async function logout() {
    await fetch("/api/session/logout", { method: "POST" });
    router.push("/login");
    router.refresh();
  }

  return (
    <div className="workspace">
      <aside className="sidebar">
        <Link className="brand" href="/dashboard"><span className="brand-mark">S</span> SafeScope</Link>
        <nav className="nav" aria-label="Primary navigation">
          {links.map(({ href, label, icon: Icon }) => (
            <Link key={href} href={href} className={`nav-link ${pathname.startsWith(href) ? "active" : ""}`}>
              <Icon size={17} /> {label}
            </Link>
          ))}
          {user?.role === "admin" && (
            <Link href="/team" className={`nav-link ${pathname.startsWith("/team") ? "active" : ""}`}>
              <Users size={17} /> Team
            </Link>
          )}
        </nav>
        <div className="sidebar-foot"><small>AI Security Workspace</small></div>
      </aside>
      <main className="main">
        <header className="topbar">
          <div className="topbar-title">{pathname.startsWith("/projects") ? "Asset Management" : "Security Overview"}</div>
          <div className="user-menu">
            <div className="avatar">{user?.name?.slice(0, 1).toUpperCase() ?? "S"}</div>
            <div className="user-meta"><span>{user?.name ?? "Operator"}</span><span>{user?.role ?? "loading"}</span></div>
            <button className="icon-plain" onClick={logout} title="Sign out" aria-label="Sign out"><LogOut size={17} /></button>
          </div>
        </header>
        {children}
      </main>
      <nav className="mobile-nav" aria-label="Mobile navigation">
        {links.map(({ href, label, icon: Icon }) => (
          <Link key={href} href={href} className={pathname.startsWith(href) ? "active" : ""}>
            <Icon size={18} /> {label}
          </Link>
        ))}
      </nav>
    </div>
  );
}
