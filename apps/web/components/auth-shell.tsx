import { ShieldCheck } from "lucide-react";

export function AuthShell({ children }: { children: React.ReactNode }) {
  return (
    <main className="auth-shell">
      <section className="auth-context">
        <div className="brand"><span className="brand-mark">S</span> SafeScope</div>
        <div>
          <h1>See your exposure. Act with confidence.</h1>
          <p>One operational workspace for projects, internet-facing assets, scanner intelligence, and AI-assisted security analysis.</p>
        </div>
        <div className="auth-points">
          <span><ShieldCheck size={14} /> JWT secured</span>
          <span>Role-aware access</span>
          <span>Built for automation</span>
        </div>
      </section>
      <section className="auth-panel">{children}</section>
    </main>
  );
}
