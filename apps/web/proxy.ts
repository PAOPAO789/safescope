import { NextRequest, NextResponse } from "next/server";
import { SESSION_COOKIE } from "@/lib/server-api";

export function proxy(request: NextRequest) {
  const authenticated = Boolean(request.cookies.get(SESSION_COOKIE)?.value);
  const isAuthPage = request.nextUrl.pathname === "/login" || request.nextUrl.pathname === "/register";
  if (!authenticated && !isAuthPage) {
    return NextResponse.redirect(new URL("/login", request.url));
  }
  if (authenticated && isAuthPage) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/projects/:path*", "/team/:path*", "/login", "/register"],
};
