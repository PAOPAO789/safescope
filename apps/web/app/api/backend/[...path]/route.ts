import { NextRequest } from "next/server";
import { proxyRequest } from "@/lib/server-api";

type Context = { params: Promise<{ path: string[] }> };

async function forward(request: NextRequest, context: Context) {
  return proxyRequest(request, (await context.params).path);
}

export const GET = forward;
export const POST = forward;
export const PUT = forward;
export const DELETE = forward;
