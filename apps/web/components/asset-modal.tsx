"use client";

import { FormEvent, useState } from "react";
import { X } from "lucide-react";
import { api } from "@/lib/api";
import type { Asset } from "@/lib/types";

export function AssetModal({
  projectId,
  asset,
  onClose,
  onSaved,
}: {
  projectId: string;
  asset?: Asset;
  onClose: () => void;
  onSaved: (asset: Asset) => void;
}) {
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    const form = new FormData(event.currentTarget);
    const body = {
      type: form.get("type"),
      value: form.get("value"),
      status: form.get("status"),
      tags: String(form.get("tags") ?? "").split(",").map((tag) => tag.trim()).filter(Boolean),
      metadata: { source: "manual" },
    };
    try {
      const path = asset ? `/assets/${asset.id}` : `/projects/${projectId}/assets`;
      const saved = await api<Asset>(path, { method: asset ? "PUT" : "POST", body: JSON.stringify(body) });
      onSaved(saved);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to save asset");
      setLoading(false);
    }
  }

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <div className="modal" role="dialog" aria-modal="true" aria-labelledby="asset-modal-title">
        <div className="modal-head"><h2 id="asset-modal-title">{asset ? "Edit asset" : "Add asset"}</h2><button className="icon-plain" onClick={onClose} aria-label="Close"><X size={19} /></button></div>
        <form className="modal-body" onSubmit={submit}>
          {!asset && <>
            <div className="field"><label htmlFor="asset-type">Type</label><select id="asset-type" name="type" defaultValue="domain"><option value="domain">Domain</option><option value="ip">IP address</option><option value="url">URL</option><option value="service">Service</option></select></div>
            <div className="field"><label htmlFor="asset-value">Value</label><input id="asset-value" name="value" placeholder="example.com" required autoFocus /></div>
          </>}
          {asset && <div className="field"><label htmlFor="asset-status">Status</label><select id="asset-status" name="status" defaultValue={asset.status}><option value="unknown">Unknown</option><option value="alive">Alive</option><option value="down">Down</option></select></div>}
          <div className="field"><label htmlFor="asset-tags">Tags</label><input id="asset-tags" name="tags" defaultValue={asset?.tags.join(", ")} placeholder="production, external" /></div>
          {error && <p className="form-error">{error}</p>}
          <div className="modal-actions"><button className="button secondary" type="button" onClick={onClose}>Cancel</button><button className="button" disabled={loading}>{asset ? "Save changes" : "Add asset"}</button></div>
        </form>
      </div>
    </div>
  );
}
