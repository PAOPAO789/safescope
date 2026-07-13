import { NextRequest } from "next/server";
import { createSession } from "@/lib/server-api";

export async function POST(request: NextRequest) {
  return createSession(request, "login");
}
