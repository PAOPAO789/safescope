"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { ArrowRight, LoaderCircle } from "lucide-react";

export function AuthForm({ mode }: { mode: "login" | "register" }) {
  const router = useRouter();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    const data = Object.fromEntries(new FormData(event.currentTarget));
    const response = await fetch(`/api/session/${mode}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    const body = await response.json().catch(() => null);
    if (!response.ok) {
      setError(body?.error?.message ?? "Unable to continue");
      setLoading(false);
      return;
    }
    router.push("/dashboard");
    router.refresh();
  }

  const registering = mode === "register";
  return (
    <form className="auth-form" onSubmit={submit}>
      <h2>{registering ? "Create your workspace" : "Welcome back"}</h2>
      <p>{registering ? "Set up your first operator account." : "Sign in to continue to your security workspace."}</p>
      {registering && (
        <div className="field">
          <label htmlFor="name">Full name</label>
          <input id="name" name="name" autoComplete="name" required />
        </div>
      )}
      <div className="field">
        <label htmlFor="email">Email</label>
        <input id="email" name="email" type="email" autoComplete="email" required />
      </div>
      <div className="field">
        <label htmlFor="password">Password</label>
        <input id="password" name="password" type="password" minLength={8} autoComplete={registering ? "new-password" : "current-password"} required />
      </div>
      {error && <p className="form-error">{error}</p>}
      <button className="button full" disabled={loading}>
        {loading ? <LoaderCircle size={16} /> : <ArrowRight size={16} />}
        {registering ? "Create account" : "Sign in"}
      </button>
      <div className="auth-switch">
        {registering ? "Already have an account? " : "New to SafeScope? "}
        <Link href={registering ? "/login" : "/register"}>{registering ? "Sign in" : "Create account"}</Link>
      </div>
    </form>
  );
}
